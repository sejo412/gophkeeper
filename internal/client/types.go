package client

import pb "github.com/sejo412/gophkeeper/proto"

func protoRecordType(value pb.RecordType) *pb.RecordType {
	return &value
}
