package client

import (
	"bufio"
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sejo412/gophkeeper/internal/models"
	"github.com/sejo412/gophkeeper/pkg/crypt"
	pb "github.com/sejo412/gophkeeper/proto"
)

func createPassword(ctx context.Context, c *Client, scanner *bufio.Scanner) {
	record := models.Password{}
	fmt.Print("Login: ")
	scanner.Scan()
	record.Login = scanner.Text()
	fmt.Print("Password: ")
	scanner.Scan()
	record.Password = scanner.Text()
	fmt.Print("Meta: ")
	scanner.Scan()
	record.Meta = models.Meta(scanner.Text())

	encoded, err := encryptPassword(c.publicKey, record)
	if err != nil {
		fmt.Println("Error encode password record: ", err)
		return
	}

	bin, err := json.Marshal(encoded)
	if err != nil {
		fmt.Println("Error marshalling record: ", err)
		return
	}

	resp, err := c.client.Create(
		ctx, &pb.AddRecordRequest{
			Type:   protoRecordType(pb.RecordType_PASSWORD),
			Record: bin,
		},
	)
	if err != nil {
		fmt.Printf("Error creating record: %v\n", err)
	} else {
		fmt.Printf("Created record: %v\n", resp)
	}
}

func getPassword(ctx context.Context, c *Client, scanner *bufio.Scanner) {
	fmt.Print("Choose password: ")
	scanner.Scan()
	num := scanner.Text()
	id, err := strconv.Atoi(num)
	if err != nil {
		fmt.Println("Invalid number: ", err)
		return
	}
	resp, err := c.client.Read(
		ctx, &pb.GetRecordRequest{
			Type:         protoRecordType(pb.RecordType_PASSWORD),
			RecordNumber: protoID(id),
		},
	)
	if err != nil {
		fmt.Printf("Error getting password: %v\n", err)
		return
	}
	record := resp.GetRecord()
	password := models.RecordEncrypted{}
	err = json.Unmarshal(record, &password)
	if err != nil {
		fmt.Printf("Error unmarshalling record: %v\n", err)
		return
	}
	res, err := decryptPassword(c.privateKey, password.Password)
	if err != nil {
		fmt.Println("Error decrypting password: ", err)
		return
	}
	fmt.Println("Login: ", res.Login)
	fmt.Println("Password: ", res.Password)
	fmt.Println("Meta: ", res.Meta)
}

func encryptPassword(key *rsa.PublicKey, password models.Password) (models.PasswordEncrypted, error) {
	var err error
	res := models.PasswordEncrypted{}
	res.Login, err = crypt.EncryptWithPublicKey(key, []byte(password.Login))
	if err != nil {
		return res, err
	}
	res.Password, err = crypt.EncryptWithPublicKey(key, []byte(password.Password))
	if err != nil {
		return res, err
	}
	res.Meta, err = crypt.EncryptWithPublicKey(key, []byte(password.Meta))
	if err != nil {
		return res, err
	}
	return res, nil
}

func decryptPassword(key *rsa.PrivateKey, password models.PasswordEncrypted) (models.Password, error) {
	res := models.Password{}
	res.ID = password.ID
	login, err := crypt.DecryptWithPrivateKey(key, password.Login)
	if err != nil {
		return res, err
	}
	res.Login = string(login)
	pwd, err := crypt.DecryptWithPrivateKey(key, password.Password)
	if err != nil {
		return res, err
	}
	res.Password = string(pwd)
	meta, err := crypt.DecryptWithPrivateKey(key, password.Meta)
	if err != nil {
		return res, err
	}
	res.Meta = models.Meta(meta)
	return res, nil
}
