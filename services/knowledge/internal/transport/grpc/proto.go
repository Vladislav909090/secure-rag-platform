package grpc

import (
	"time"

	pb "secure-rag-platform/api/gen/go/knowledge/v1"
	"secure-rag-platform/services/knowledge/internal/model"

	"google.golang.org/protobuf/types/known/structpb"
)

func documentToProto(doc *model.Document) *pb.Document {
	d := &pb.Document{
		Id:             doc.ID,
		Uuid:           doc.UUID,
		Title:          doc.Title,
		FileName:       doc.FileName,
		FileExtension:  doc.FileExtension,
		MimeType:       doc.MimeType,
		SizeBytes:      doc.SizeBytes,
		ChecksumSha256: doc.ChecksumSHA256,
		StorageKey:     doc.StorageKey,
		IndexStatus:    doc.IndexStatus,
		CreatedAt:      doc.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      doc.UpdatedAt.Format(time.RFC3339),
	}
	if doc.Description != nil {
		d.Description = *doc.Description
	}
	if doc.DeletedAt != nil {
		d.DeletedAt = doc.DeletedAt.Format(time.RFC3339)
	}
	if doc.Attributes != nil {
		d.Attributes, _ = structpb.NewStruct(doc.Attributes)
	}

	return d
}
