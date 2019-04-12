package server

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Tnze/gomcbot/CFB8"
	pk "github.com/Tnze/gomcbot/packet"
	"github.com/satori/go.uuid"
)

const verifyTokenLen = 16

func (c *Client) login() (err error) {
	c.Name, err = c.loginStart()
	if err != nil {
		return fmt.Errorf("unexpected_query_response")
	}

	if Threshold >= 0 {
		err = c.setCompression(Threshold)
		if err != nil {
			return fmt.Errorf("unexpected_query_response")
		}
	}

	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return fmt.Errorf("unexpected_query_response")
	}

	publicKey, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return fmt.Errorf("generic")
	}

	VT1, err := c.encryptionRequest(publicKey)
	if err != nil {
		return fmt.Errorf("generic")
	}

	ESharedSecret, EVerifyToken, err := c.encryptionResponse()
	if err != nil {
		return fmt.Errorf("generic")
	}

	SharedSecret, err := rsa.DecryptPKCS1v15(rand.Reader, key, ESharedSecret)
	if err != nil {
		return fmt.Errorf("generic")
	}
	VT2, err := rsa.DecryptPKCS1v15(rand.Reader, key, EVerifyToken)
	if err != nil {
		return fmt.Errorf("generic")
	}

	if !bytes.Equal(VT1, VT2) {
		return fmt.Errorf("generic")
	}

	b, err := aes.NewCipher(SharedSecret)
	if err != nil {
		return fmt.Errorf("generic")
	}
	//启用加密
	c.reader = bufio.NewReader(cipher.StreamReader{
		S: CFB8.NewCFB8Decrypt(b, SharedSecret),
		R: c.conn,
	})
	c.Writer = cipher.StreamWriter{
		S: CFB8.NewCFB8Encrypt(b, SharedSecret),
		W: c.conn,
	}

	hash := authDigest("", SharedSecret, publicKey)
	resp, err := c.authentication(hash)
	if err != nil {
		return fmt.Errorf("authservers_down")
	}

	c.ID, err = uuid.FromString(resp.ID)
	if err != nil {
		return fmt.Errorf("authservers_down")
	}

	if c.Name != resp.Name {
		return fmt.Errorf("unverified_username")
	}

	c.skin = resp.Properties[0].Value

	err = c.loginSuccess()
	if err != nil {
		return fmt.Errorf("generic")
	}
	return
}

func (c *Client) loginStart() (string, error) {
	loginStart, err := pk.RecvPacket(c.reader, c.threshold > 0)
	if err != nil {
		return "", err
	}
	if loginStart.ID != 0x00 {
		return "", fmt.Errorf("0x%02X is not LoginStart packet's ID", loginStart.ID)
	}

	return pk.UnpackString(bytes.NewReader(loginStart.Data))
}

func (c *Client) setCompression(threshold int) error {
	sc := pk.Packet{
		ID:   0x03,
		Data: pk.PackVarInt(int32(threshold)),
	}
	_, err := c.Write(sc.Pack(c.threshold))
	c.threshold = threshold
	return err
}

func (c *Client) loginSuccess() error {
	ls := pk.Packet{ID: 0x02}
	ls.Data = append(ls.Data, pk.PackString(c.ID.String())...)
	ls.Data = append(ls.Data, pk.PackString(c.Name)...)
	_, err := c.Write(ls.Pack(c.threshold))
	return err
}

func (c *Client) encryptionRequest(publicKey []byte) ([]byte, error) {
	var verifyToken [verifyTokenLen]byte
	_, err := rand.Read(verifyToken[:])
	if err != nil {
		return nil, err
	}

	er := pk.Packet{ID: 0x01}
	er.Data = append(er.Data, pk.PackString("")...)
	er.Data = append(er.Data, pk.PackVarInt(int32(len(publicKey)))...)
	er.Data = append(er.Data, publicKey...)
	er.Data = append(er.Data, pk.PackVarInt(verifyTokenLen)...)
	er.Data = append(er.Data, verifyToken[:]...)

	_, err = c.Write(er.Pack(c.threshold))
	return verifyToken[:], err
}

func (c *Client) encryptionResponse() ([]byte, []byte, error) {
	p, err := pk.RecvPacket(c.reader, c.threshold > 0)
	if err != nil {
		return nil, nil, err
	}
	if p.ID != 0x01 {
		return nil, nil, fmt.Errorf("0x%02X is not Encryption Response", p.ID)
	}

	r := bytes.NewReader(p.Data)
	SharedSecretLength, err := pk.UnpackVarInt(r)
	if err != nil {
		return nil, nil, err
	}
	SharedSecret, err := pk.ReadNBytes(r, int(SharedSecretLength))
	if err != nil {
		return nil, nil, err
	}
	VerifyTokenLength, err := pk.UnpackVarInt(r)
	if err != nil {
		return nil, nil, err
	}
	VerifyToken, err := pk.ReadNBytes(r, int(VerifyTokenLength))
	if err != nil {
		return nil, nil, err
	}

	return SharedSecret, VerifyToken, nil
}

type authResp struct {
	ID, Name   string
	Properties [1]struct {
		Name, Value, Signature string
	}
}

func (c *Client) authentication(hash string) (*authResp, error) {
	resp, err := http.Get(fmt.Sprintf("https://sessionserver.mojang.com/session/minecraft/hasJoined?username=%s&serverId=%s",
		c.Name, hash))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var Resp authResp

	err = json.Unmarshal(body, &Resp)
	if err != nil {
		return nil, err
	}

	return &Resp, nil
}

// authDigest computes a special SHA-1 digest required for Minecraft web
// authentication on Premium servers (online-mode=true).
// Source: http://wiki.vg/Protocol_Encryption#Server
//
// Also many, many thanks to SirCmpwn and his wonderful gist (C#):
// https://gist.github.com/SirCmpwn/404223052379e82f91e6
func authDigest(serverID string, sharedSecret, publicKey []byte) string {
	h := sha1.New()
	h.Write([]byte(serverID))
	h.Write(sharedSecret)
	h.Write(publicKey)
	hash := h.Sum(nil)

	// Check for negative hashes
	negative := (hash[0] & 0x80) == 0x80
	if negative {
		hash = twosComplement(hash)
	}

	// Trim away zeroes
	res := strings.TrimLeft(fmt.Sprintf("%x", hash), "0")
	if negative {
		res = "-" + res
	}

	return res
}

// little endian
func twosComplement(p []byte) []byte {
	carry := true
	for i := len(p) - 1; i >= 0; i-- {
		p[i] = byte(^p[i])
		if carry {
			carry = p[i] == 0xff
			p[i]++
		}
	}
	return p
}
