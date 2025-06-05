package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sejo412/gophkeeper/internal/models"
	"github.com/sejo412/gophkeeper/pkg/crypt"
	pb "github.com/sejo412/gophkeeper/proto"
)

func createRecord(ctx context.Context, c *Client, t models.RecordType, scanner *bufio.Scanner) {
	encrypted := models.RecordEncrypted{
		Password: models.PasswordEncrypted{},
		Text:     models.TextEncrypted{},
		Bin:      models.BinEncrypted{},
		Bank:     models.BankEncrypted{},
	}
	f := fields(t)
	for _, field := range f {
		fmt.Printf("%s: ", field.String())
		scanner.Scan()
		val := scanner.Text()
		valEnc, err := crypt.EncryptWithPublicKey(c.publicKey, []byte(val))
		if err != nil {
			fmt.Println(err)
			return
		}
		switch t {
		case models.RecordPassword:
			switch field {
			case FieldLogin:
				encrypted.Password.Login = valEnc
			case FieldPassword:
				encrypted.Password.Password = valEnc
			case FieldMeta:
				encrypted.Password.Meta = valEnc
			default:
			}
		case models.RecordText:
			switch field {
			case FieldText:
				encrypted.Text.Text = valEnc
			case FieldMeta:
				encrypted.Text.Meta = valEnc
			default:
			}
		case models.RecordBin:
			switch field {
			case FieldData:
				encrypted.Bin.Data = valEnc
			case FieldMeta:
				encrypted.Bin.Meta = valEnc
			default:
			}
		case models.RecordBank:
			switch field {
			case FieldNumber:
				encrypted.Bank.Number = valEnc
			case FieldDate:
				encrypted.Bank.Date = valEnc
			case FieldCVV:
				encrypted.Bank.Cvv = valEnc
			case FieldMeta:
				encrypted.Bank.Meta = valEnc
			default:
			}
		default:
		}
	}
	bin, err := json.Marshal(&encrypted)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = c.client.Create(
		ctx, &pb.AddRecordRequest{
			Type:   protoRecordType(modelRecordTypeToProto(t)),
			Record: bin,
		},
	)
	if err != nil {
		fmt.Printf("failed to send request: %v\n", err)
	}
}

func readRecord(ctx context.Context, c *Client, t models.RecordType, scanner *bufio.Scanner) {
	fmt.Printf("Choose %s: ", t.String())
	scanner.Scan()
	val := scanner.Text()
	id, err := strconv.Atoi(val)
	if err != nil {
		fmt.Println("Invalid ID: ", err)
		return
	}
	resp, err := c.client.Read(
		ctx,
		&pb.GetRecordRequest{
			Type:         protoRecordType(modelRecordTypeToProto(t)),
			RecordNumber: protoID(id),
		},
	)
	if err != nil {
		fmt.Printf("Error getting %s: %v\n", t.String(), err)
		return
	}
	protoRecord := resp.GetRecord()
	record := models.RecordEncrypted{}
	err = json.Unmarshal(protoRecord, &record)
	if err != nil {
		fmt.Printf("Error unmarshalling record: %v\n", err)
		return
	}
	for _, field := range fields(t) {
		var valDec []byte
		var err error
		switch t {
		case models.RecordPassword:
			switch field {
			case FieldLogin:
				valDec, err = crypt.DecryptWithPrivateKey(c.privateKey, record.Password.Login)
			case FieldPassword:
				valDec, err = crypt.DecryptWithPrivateKey(c.privateKey, record.Password.Password)
			case FieldMeta:
				valDec, err = crypt.DecryptWithPrivateKey(c.privateKey, record.Password.Meta)
			default:
			}
		case models.RecordText:
			switch field {
			case FieldText:
				valDec, err = crypt.DecryptWithPrivateKey(c.privateKey, record.Text.Text)
			case FieldMeta:
				valDec, err = crypt.DecryptWithPrivateKey(c.privateKey, record.Text.Meta)
			default:
			}
		case models.RecordBin:
			switch field {
			case FieldData:
				valDec, err = crypt.DecryptWithPrivateKey(c.privateKey, record.Bin.Data)
			case FieldMeta:
				valDec, err = crypt.DecryptWithPrivateKey(c.privateKey, record.Bin.Meta)
			default:
			}
		case models.RecordBank:
			switch field {
			case FieldNumber:
				valDec, err = crypt.DecryptWithPrivateKey(c.privateKey, record.Bank.Number)
			case FieldDate:
				valDec, err = crypt.DecryptWithPrivateKey(c.privateKey, record.Bank.Date)
			case FieldCVV:
				valDec, err = crypt.DecryptWithPrivateKey(c.privateKey, record.Bank.Cvv)
			case FieldMeta:
				valDec, err = crypt.DecryptWithPrivateKey(c.privateKey, record.Bank.Meta)
			default:
			}
		default:
		}
		if err != nil {
			fmt.Printf("Error decrypt %s: %v\n", field.String(), err)
			return
		}
		fmt.Printf("%s: %s\n", field.String(), string(valDec))
	}
}

func listRecords(ctx context.Context, c *Client, t models.RecordType) {
	resp, err := c.client.List(
		ctx, &pb.ListRequest{
			Type: protoRecordType(modelRecordTypeToProto(t)),
		},
	)
	if err != nil {
		fmt.Printf("Error listing records: %v\n", err)
		return
	}
	records := resp.GetRecords()
	data := models.RecordsEncrypted{}
	if err = json.Unmarshal(records, &data); err != nil {
		fmt.Printf("Error unmarshalling records: %v\n", err)
		return
	}
	resultEnc := make(map[models.ID][]byte)
	switch t {
	case models.RecordPassword:
		for _, record := range data.Password {
			resultEnc[record.ID] = record.Meta
		}
	case models.RecordText:
		for _, record := range data.Text {
			resultEnc[record.ID] = record.Meta
		}
	case models.RecordBin:
		for _, record := range data.Bin {
			resultEnc[record.ID] = record.Meta
		}
	case models.RecordBank:
		for _, record := range data.Bank {
			resultEnc[record.ID] = record.Meta
		}
	default:
	}
	for id, record := range resultEnc {
		decrypted, err := crypt.DecryptWithPrivateKey(c.privateKey, record)
		if err != nil {
			fmt.Printf("Error decrypting record: %v\n", err)
			return
		}
		fmt.Printf("%d: %s\n", id, string(decrypted))
	}
}
