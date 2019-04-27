// Code generated by protoc-gen-go. DO NOT EDIT.
// source: currencyrates.proto

package currencyrates

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type GetCurrentRateRequest struct {
	//@inject_tag: validate:"required,alpha,len=3"
	From string `protobuf:"bytes,1,opt,name=from,proto3" json:"from,omitempty" validate:"required,alpha,len=3"`
	//@inject_tag: validate:"required,alpha,len=3"
	To                   string   `protobuf:"bytes,2,opt,name=to,proto3" json:"to,omitempty" validate:"required,alpha,len=3"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-" structure:"-" validate:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-" structure:"-" validate:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-" structure:"-" validate:"-"`
}

func (m *GetCurrentRateRequest) Reset()         { *m = GetCurrentRateRequest{} }
func (m *GetCurrentRateRequest) String() string { return proto.CompactTextString(m) }
func (*GetCurrentRateRequest) ProtoMessage()    {}
func (*GetCurrentRateRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_d90eccbc6e715cb5, []int{0}
}

func (m *GetCurrentRateRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetCurrentRateRequest.Unmarshal(m, b)
}
func (m *GetCurrentRateRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetCurrentRateRequest.Marshal(b, m, deterministic)
}
func (m *GetCurrentRateRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetCurrentRateRequest.Merge(m, src)
}
func (m *GetCurrentRateRequest) XXX_Size() int {
	return xxx_messageInfo_GetCurrentRateRequest.Size(m)
}
func (m *GetCurrentRateRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetCurrentRateRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetCurrentRateRequest proto.InternalMessageInfo

func (m *GetCurrentRateRequest) GetFrom() string {
	if m != nil {
		return m.From
	}
	return ""
}

func (m *GetCurrentRateRequest) GetTo() string {
	if m != nil {
		return m.To
	}
	return ""
}

type GetCentralBankRateRequest struct {
	//@inject_tag: validate:"required,alpha,len=3"
	From string `protobuf:"bytes,1,opt,name=from,proto3" json:"from,omitempty" validate:"required,alpha,len=3"`
	//@inject_tag: validate:"required,alpha,len=3"
	To string `protobuf:"bytes,2,opt,name=to,proto3" json:"to,omitempty" validate:"required,alpha,len=3"`
	//@inject_tag: validate:"required"
	Datetime             *timestamp.Timestamp `protobuf:"bytes,3,opt,name=datetime,proto3" json:"datetime,omitempty" validate:"required"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-" bson:"-" structure:"-" validate:"-"`
	XXX_unrecognized     []byte               `json:"-" bson:"-" structure:"-" validate:"-"`
	XXX_sizecache        int32                `json:"-" bson:"-" structure:"-" validate:"-"`
}

func (m *GetCentralBankRateRequest) Reset()         { *m = GetCentralBankRateRequest{} }
func (m *GetCentralBankRateRequest) String() string { return proto.CompactTextString(m) }
func (*GetCentralBankRateRequest) ProtoMessage()    {}
func (*GetCentralBankRateRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_d90eccbc6e715cb5, []int{1}
}

func (m *GetCentralBankRateRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetCentralBankRateRequest.Unmarshal(m, b)
}
func (m *GetCentralBankRateRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetCentralBankRateRequest.Marshal(b, m, deterministic)
}
func (m *GetCentralBankRateRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetCentralBankRateRequest.Merge(m, src)
}
func (m *GetCentralBankRateRequest) XXX_Size() int {
	return xxx_messageInfo_GetCentralBankRateRequest.Size(m)
}
func (m *GetCentralBankRateRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetCentralBankRateRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetCentralBankRateRequest proto.InternalMessageInfo

func (m *GetCentralBankRateRequest) GetFrom() string {
	if m != nil {
		return m.From
	}
	return ""
}

func (m *GetCentralBankRateRequest) GetTo() string {
	if m != nil {
		return m.To
	}
	return ""
}

func (m *GetCentralBankRateRequest) GetDatetime() *timestamp.Timestamp {
	if m != nil {
		return m.Datetime
	}
	return nil
}

type RateData struct {
	//@inject_tag: validate:"required,hexadecimal,len=24" json:"id" bson:"_id"
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id" validate:"required,hexadecimal,len=24" bson:"_id"`
	//@inject_tag: validate:"required,alpha,len=6" json:"created_at" bson:"created_at"
	CreatedAt *timestamp.Timestamp `protobuf:"bytes,2,opt,name=created_at,json=createdAt,proto3" json:"created_at" validate:"required,alpha,len=6" bson:"created_at"`
	//@inject_tag: validate:"required,alpha,len=6" json:"pair" bson:"pair"
	Pair string `protobuf:"bytes,3,opt,name=pair,proto3" json:"pair" validate:"required,alpha,len=6" bson:"pair"`
	//@inject_tag: validate:"required,numeric,gt=0" json:"rate" bson:"rate"
	Rate float64 `protobuf:"fixed64,4,opt,name=rate,proto3" json:"rate" validate:"required,numeric,gt=0" bson:"rate"`
	//@inject_tag: validate:"required,numeric,gt=0" json:"correction" bson:"correction"
	Correction float64 `protobuf:"fixed64,5,opt,name=correction,proto3" json:"correction" validate:"required,numeric,gt=0" bson:"correction"`
	//@inject_tag: validate:"required,numeric,gt=0" json:"corrected_rate" bson:"corrected_rate"
	CorrectedRate float64 `protobuf:"fixed64,6,opt,name=corrected_rate,json=correctedRate,proto3" json:"corrected_rate" validate:"required,numeric,gt=0" bson:"corrected_rate"`
	//@inject_tag: json:"is_cb_rate" bson:"is_cb_rate"
	IsCbRate bool `protobuf:"varint,7,opt,name=is_cb_rate,json=isCbRate,proto3" json:"is_cb_rate" bson:"is_cb_rate"`
	//@inject_tag: validate:"required,alpha" json:"source" bson:"source"
	Source               string   `protobuf:"bytes,8,opt,name=source,proto3" json:"source" validate:"required,alpha" bson:"source"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-" structure:"-" validate:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-" structure:"-" validate:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-" structure:"-" validate:"-"`
}

func (m *RateData) Reset()         { *m = RateData{} }
func (m *RateData) String() string { return proto.CompactTextString(m) }
func (*RateData) ProtoMessage()    {}
func (*RateData) Descriptor() ([]byte, []int) {
	return fileDescriptor_d90eccbc6e715cb5, []int{2}
}

func (m *RateData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RateData.Unmarshal(m, b)
}
func (m *RateData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RateData.Marshal(b, m, deterministic)
}
func (m *RateData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RateData.Merge(m, src)
}
func (m *RateData) XXX_Size() int {
	return xxx_messageInfo_RateData.Size(m)
}
func (m *RateData) XXX_DiscardUnknown() {
	xxx_messageInfo_RateData.DiscardUnknown(m)
}

var xxx_messageInfo_RateData proto.InternalMessageInfo

func (m *RateData) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *RateData) GetCreatedAt() *timestamp.Timestamp {
	if m != nil {
		return m.CreatedAt
	}
	return nil
}

func (m *RateData) GetPair() string {
	if m != nil {
		return m.Pair
	}
	return ""
}

func (m *RateData) GetRate() float64 {
	if m != nil {
		return m.Rate
	}
	return 0
}

func (m *RateData) GetCorrection() float64 {
	if m != nil {
		return m.Correction
	}
	return 0
}

func (m *RateData) GetCorrectedRate() float64 {
	if m != nil {
		return m.CorrectedRate
	}
	return 0
}

func (m *RateData) GetIsCbRate() bool {
	if m != nil {
		return m.IsCbRate
	}
	return false
}

func (m *RateData) GetSource() string {
	if m != nil {
		return m.Source
	}
	return ""
}

func init() {
	proto.RegisterType((*GetCurrentRateRequest)(nil), "currencyrates.GetCurrentRateRequest")
	proto.RegisterType((*GetCentralBankRateRequest)(nil), "currencyrates.GetCentralBankRateRequest")
	proto.RegisterType((*RateData)(nil), "currencyrates.RateData")
}

func init() { proto.RegisterFile("currencyrates.proto", fileDescriptor_d90eccbc6e715cb5) }

var fileDescriptor_d90eccbc6e715cb5 = []byte{
	// 362 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x52, 0x4d, 0x4f, 0xe3, 0x30,
	0x10, 0x5d, 0x67, 0xbb, 0xdd, 0x74, 0x56, 0xed, 0xc1, 0xfb, 0x95, 0xad, 0x56, 0xbb, 0x51, 0x04,
	0x52, 0x4e, 0xa9, 0x54, 0x24, 0x24, 0xc4, 0x09, 0x5a, 0xc1, 0x99, 0xc0, 0xbd, 0x38, 0xce, 0xb4,
	0xb2, 0x68, 0xe3, 0xe2, 0x4c, 0x40, 0xfc, 0x46, 0xfe, 0x11, 0x27, 0x64, 0x37, 0xad, 0x08, 0x8a,
	0x2a, 0x71, 0x1b, 0xbf, 0x79, 0xf3, 0xe6, 0xc9, 0x6f, 0xe0, 0xbb, 0xac, 0x8c, 0xc1, 0x42, 0x3e,
	0x19, 0x41, 0x58, 0x26, 0x6b, 0xa3, 0x49, 0xf3, 0x7e, 0x03, 0x1c, 0xfe, 0x5f, 0x68, 0xbd, 0x58,
	0xe2, 0xc8, 0x35, 0xb3, 0x6a, 0x3e, 0x22, 0xb5, 0xc2, 0x92, 0xc4, 0x6a, 0xbd, 0xe1, 0x47, 0xa7,
	0xf0, 0xf3, 0x12, 0x69, 0xe2, 0x86, 0x28, 0x15, 0x84, 0x29, 0xde, 0x57, 0x58, 0x12, 0xe7, 0xd0,
	0x99, 0x1b, 0xbd, 0x0a, 0x58, 0xc8, 0xe2, 0x5e, 0xea, 0x6a, 0x3e, 0x00, 0x8f, 0x74, 0xe0, 0x39,
	0xc4, 0x23, 0x1d, 0x3d, 0xc2, 0x1f, 0x3b, 0x8c, 0x05, 0x19, 0xb1, 0x3c, 0x17, 0xc5, 0xdd, 0x07,
	0x05, 0xf8, 0x31, 0xf8, 0xb9, 0x20, 0xb4, 0xa6, 0x82, 0xcf, 0x21, 0x8b, 0xbf, 0x8d, 0x87, 0xc9,
	0xc6, 0x71, 0xb2, 0x75, 0x9c, 0xdc, 0x6c, 0x1d, 0xa7, 0x3b, 0x6e, 0xf4, 0xc2, 0xc0, 0xb7, 0xbb,
	0xa6, 0x82, 0x84, 0x15, 0x55, 0x79, 0xbd, 0xc6, 0x53, 0x39, 0x3f, 0x01, 0x90, 0x06, 0x05, 0x61,
	0x3e, 0x13, 0xe4, 0x96, 0xed, 0x97, 0xed, 0xd5, 0xec, 0x33, 0xe7, 0x79, 0x2d, 0x94, 0x71, 0x5e,
	0x7a, 0xa9, 0xab, 0x2d, 0x66, 0xff, 0x32, 0xe8, 0x84, 0x2c, 0x66, 0xa9, 0xab, 0xf9, 0x3f, 0x00,
	0xa9, 0x8d, 0x41, 0x49, 0x4a, 0x17, 0xc1, 0x17, 0xd7, 0x79, 0x83, 0xf0, 0x43, 0x18, 0xd4, 0x2f,
	0xcc, 0x67, 0x6e, 0xba, 0xeb, 0x38, 0xfd, 0x1d, 0x6a, 0xdd, 0xf3, 0xbf, 0x00, 0xaa, 0x9c, 0xc9,
	0x6c, 0x43, 0xf9, 0x1a, 0xb2, 0xd8, 0x4f, 0x7d, 0x55, 0x4e, 0x32, 0xd7, 0xfd, 0x05, 0xdd, 0x52,
	0x57, 0x46, 0x62, 0xe0, 0x3b, 0x3b, 0xf5, 0x6b, 0xfc, 0xcc, 0xe0, 0x47, 0x23, 0xe5, 0x6b, 0x34,
	0x0f, 0x4a, 0x22, 0xbf, 0x82, 0x41, 0x33, 0x4b, 0x7e, 0x90, 0x34, 0x6f, 0xa4, 0x35, 0xea, 0xe1,
	0xef, 0x77, 0xac, 0xed, 0xcf, 0x46, 0x9f, 0xf8, 0x6d, 0x5b, 0xc2, 0x17, 0xda, 0x4c, 0xad, 0x7a,
	0xdc, 0xa2, 0xde, 0x7a, 0x0b, 0x7b, 0x36, 0x64, 0x5d, 0x97, 0xc8, 0xd1, 0x6b, 0x00, 0x00, 0x00,
	0xff, 0xff, 0x80, 0x71, 0x69, 0x5e, 0xce, 0x02, 0x00, 0x00,
}
