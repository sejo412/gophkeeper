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
	bin, err := writeRecord(ctx, c, t, scanner)
	if err != nil {
		fmt.Printf("error writing record: %v\n", err)
		return
	}
	_, err = c.client.Create(
		ctx, &pb.AddRecordRequest{
			Type:   protoRecordType(modelRecordTypeToProto(t)),
			Record: bin,
		},
	)
	if err != nil {
		fmt.Printf("failed create %s: %v\n", t.String(), err)
	}
	fmt.Printf("Created %s\n", t.String())
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

func updateRecord(ctx context.Context, c *Client, t models.RecordType, scanner *bufio.Scanner) {
	fmt.Printf("Choose %s: ", t.String())
	scanner.Scan()
	val := scanner.Text()
	id, err := strconv.Atoi(val)
	if err != nil {
		fmt.Println("Invalid ID: ", err)
		return
	}
	bin, err := writeRecord(ctx, c, t, scanner)
	if err != nil {
		fmt.Printf("error writing record: %v\n", err)
		return
	}
	_, err = c.client.Update(
		ctx, &pb.UpdateRecordRequest{
			Type:         protoRecordType(modelRecordTypeToProto(t)),
			RecordNumber: protoID(id),
			Record:       bin,
		},
	)
	if err != nil {
		fmt.Printf("failed update %s with ID %d: %v\n", t.String(), id, err)
	}
	fmt.Printf("Updated %s: %d\n", t.String(), id)
}

func deleteRecord(ctx context.Context, c *Client, t models.RecordType, scanner *bufio.Scanner) {
	fmt.Printf("Choose %s: ", t.String())
	scanner.Scan()
	val := scanner.Text()
	id, err := strconv.Atoi(val)
	if err != nil {
		fmt.Println("Invalid ID: ", err)
		return
	}
	_, err = c.client.Delete(
		ctx, &pb.DeleteRecordRequest{
			Type:         protoRecordType(modelRecordTypeToProto(t)),
			RecordNumber: protoID(id),
		},
	)
	if err != nil {
		fmt.Printf("Failed deleting %s with ID %d: %v\n", t.String(), id, err)
	}
	fmt.Printf("Deleted %s: %d\n", t.String(), id)
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

func listAllRecords(ctx context.Context, c *Client) {
	clearScreen()
	resp, err := c.client.ListAll(ctx, nil)
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
	fmt.Printf("%s:\n", models.RecordPassword.String())
	for _, record := range data.Password {
		valDec, err := crypt.DecryptWithPrivateKey(c.privateKey, record.Meta)
		if err != nil {
			fmt.Printf("Error decrypting record %d: %v\n", record.ID, err)
		} else {
			fmt.Printf("%d: %s\n", record.ID, string(valDec))
		}
	}
	fmt.Printf("%s:\n", models.RecordText.String())
	for _, record := range data.Text {
		valDec, err := crypt.DecryptWithPrivateKey(c.privateKey, record.Meta)
		if err != nil {
			fmt.Printf("Error decrypting record %d: %v\n", record.ID, err)
		} else {
			fmt.Printf("%d: %s\n", record.ID, string(valDec))
		}
	}
	fmt.Printf("%s:\n", models.RecordBin.String())
	for _, record := range data.Bin {
		valDec, err := crypt.DecryptWithPrivateKey(c.privateKey, record.Meta)
		if err != nil {
			fmt.Printf("Error decrypting record %d: %v\n", record.ID, err)
		} else {
			fmt.Printf("%d: %s\n", record.ID, string(valDec))
		}
	}
	fmt.Printf("%s:\n", models.RecordBank.String())
	for _, record := range data.Bank {
		valDec, err := crypt.DecryptWithPrivateKey(c.privateKey, record.Meta)
		if err != nil {
			fmt.Printf("Error decrypting record %d: %v\n", record.ID, err)
		} else {
			fmt.Printf("%d: %s\n", record.ID, string(valDec))
		}
	}
	waitForEnter()
}

func writeRecord(_ context.Context, c *Client, t models.RecordType, scanner *bufio.Scanner) (
	record []byte, err error,
) {
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
			return nil, err
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
		return nil, err
	}
	return bin, nil
}
