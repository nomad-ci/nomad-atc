// Code generated by protoc-gen-gogo.
// source: log.proto
// DO NOT EDIT!

/*
	Package nomadatc is a generated protocol buffer package.

	It is generated from these files:
		log.proto

	It has these top-level messages:
		OutputData
		Actions
		FileRequest
		FileData
*/
package nomadatc

import proto "github.com/gogo/protobuf/proto"
import fmt "fmt"
import math "math"

import bytes "bytes"

import strings "strings"
import github_com_gogo_protobuf_proto "github.com/gogo/protobuf/proto"
import sort "sort"
import strconv "strconv"
import reflect "reflect"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type OutputData struct {
	Stream         int64  `protobuf:"varint,1,opt,name=stream,proto3" json:"stream,omitempty"`
	Stderr         bool   `protobuf:"varint,2,opt,name=stderr,proto3" json:"stderr,omitempty"`
	Data           []byte `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
	Finished       bool   `protobuf:"varint,4,opt,name=finished,proto3" json:"finished,omitempty"`
	FinishedStatus int32  `protobuf:"varint,5,opt,name=finished_status,json=finishedStatus,proto3" json:"finished_status,omitempty"`
}

func (m *OutputData) Reset()                    { *m = OutputData{} }
func (*OutputData) ProtoMessage()               {}
func (*OutputData) Descriptor() ([]byte, []int) { return fileDescriptorLog, []int{0} }

type Actions struct {
	Signal     int64  `protobuf:"varint,1,opt,name=signal,proto3" json:"signal,omitempty"`
	Input      []byte `protobuf:"bytes,2,opt,name=input,proto3" json:"input,omitempty"`
	CloseInput bool   `protobuf:"varint,3,opt,name=close_input,json=closeInput,proto3" json:"close_input,omitempty"`
}

func (m *Actions) Reset()                    { *m = Actions{} }
func (*Actions) ProtoMessage()               {}
func (*Actions) Descriptor() ([]byte, []int) { return fileDescriptorLog, []int{1} }

type FileRequest struct {
	Path string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	Key  int32  `protobuf:"varint,2,opt,name=key,proto3" json:"key,omitempty"`
}

func (m *FileRequest) Reset()                    { *m = FileRequest{} }
func (*FileRequest) ProtoMessage()               {}
func (*FileRequest) Descriptor() ([]byte, []int) { return fileDescriptorLog, []int{2} }

type FileData struct {
	Key  int32  `protobuf:"varint,1,opt,name=key,proto3" json:"key,omitempty"`
	Data []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *FileData) Reset()                    { *m = FileData{} }
func (*FileData) ProtoMessage()               {}
func (*FileData) Descriptor() ([]byte, []int) { return fileDescriptorLog, []int{3} }

func init() {
	proto.RegisterType((*OutputData)(nil), "nomadatc.OutputData")
	proto.RegisterType((*Actions)(nil), "nomadatc.Actions")
	proto.RegisterType((*FileRequest)(nil), "nomadatc.FileRequest")
	proto.RegisterType((*FileData)(nil), "nomadatc.FileData")
}
func (this *OutputData) Equal(that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*OutputData)
	if !ok {
		that2, ok := that.(OutputData)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		if this == nil {
			return true
		}
		return false
	} else if this == nil {
		return false
	}
	if this.Stream != that1.Stream {
		return false
	}
	if this.Stderr != that1.Stderr {
		return false
	}
	if !bytes.Equal(this.Data, that1.Data) {
		return false
	}
	if this.Finished != that1.Finished {
		return false
	}
	if this.FinishedStatus != that1.FinishedStatus {
		return false
	}
	return true
}
func (this *Actions) Equal(that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*Actions)
	if !ok {
		that2, ok := that.(Actions)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		if this == nil {
			return true
		}
		return false
	} else if this == nil {
		return false
	}
	if this.Signal != that1.Signal {
		return false
	}
	if !bytes.Equal(this.Input, that1.Input) {
		return false
	}
	if this.CloseInput != that1.CloseInput {
		return false
	}
	return true
}
func (this *FileRequest) Equal(that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*FileRequest)
	if !ok {
		that2, ok := that.(FileRequest)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		if this == nil {
			return true
		}
		return false
	} else if this == nil {
		return false
	}
	if this.Path != that1.Path {
		return false
	}
	if this.Key != that1.Key {
		return false
	}
	return true
}
func (this *FileData) Equal(that interface{}) bool {
	if that == nil {
		if this == nil {
			return true
		}
		return false
	}

	that1, ok := that.(*FileData)
	if !ok {
		that2, ok := that.(FileData)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		if this == nil {
			return true
		}
		return false
	} else if this == nil {
		return false
	}
	if this.Key != that1.Key {
		return false
	}
	if !bytes.Equal(this.Data, that1.Data) {
		return false
	}
	return true
}
func (this *OutputData) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 9)
	s = append(s, "&nomadatc.OutputData{")
	s = append(s, "Stream: "+fmt.Sprintf("%#v", this.Stream)+",\n")
	s = append(s, "Stderr: "+fmt.Sprintf("%#v", this.Stderr)+",\n")
	s = append(s, "Data: "+fmt.Sprintf("%#v", this.Data)+",\n")
	s = append(s, "Finished: "+fmt.Sprintf("%#v", this.Finished)+",\n")
	s = append(s, "FinishedStatus: "+fmt.Sprintf("%#v", this.FinishedStatus)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *Actions) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 7)
	s = append(s, "&nomadatc.Actions{")
	s = append(s, "Signal: "+fmt.Sprintf("%#v", this.Signal)+",\n")
	s = append(s, "Input: "+fmt.Sprintf("%#v", this.Input)+",\n")
	s = append(s, "CloseInput: "+fmt.Sprintf("%#v", this.CloseInput)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *FileRequest) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 6)
	s = append(s, "&nomadatc.FileRequest{")
	s = append(s, "Path: "+fmt.Sprintf("%#v", this.Path)+",\n")
	s = append(s, "Key: "+fmt.Sprintf("%#v", this.Key)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func (this *FileData) GoString() string {
	if this == nil {
		return "nil"
	}
	s := make([]string, 0, 6)
	s = append(s, "&nomadatc.FileData{")
	s = append(s, "Key: "+fmt.Sprintf("%#v", this.Key)+",\n")
	s = append(s, "Data: "+fmt.Sprintf("%#v", this.Data)+",\n")
	s = append(s, "}")
	return strings.Join(s, "")
}
func valueToGoStringLog(v interface{}, typ string) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("func(v %v) *%v { return &v } ( %#v )", typ, typ, pv)
}
func extensionToGoStringLog(m github_com_gogo_protobuf_proto.Message) string {
	e := github_com_gogo_protobuf_proto.GetUnsafeExtensionsMap(m)
	if e == nil {
		return "nil"
	}
	s := "proto.NewUnsafeXXX_InternalExtensions(map[int32]proto.Extension{"
	keys := make([]int, 0, len(e))
	for k := range e {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	ss := []string{}
	for _, k := range keys {
		ss = append(ss, strconv.Itoa(k)+": "+e[int32(k)].GoString())
	}
	s += strings.Join(ss, ",") + "})"
	return s
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Task service

type TaskClient interface {
	// Sends a greeting
	EmitOutput(ctx context.Context, opts ...grpc.CallOption) (Task_EmitOutputClient, error)
	ProvideFiles(ctx context.Context, opts ...grpc.CallOption) (Task_ProvideFilesClient, error)
}

type taskClient struct {
	cc *grpc.ClientConn
}

func NewTaskClient(cc *grpc.ClientConn) TaskClient {
	return &taskClient{cc}
}

func (c *taskClient) EmitOutput(ctx context.Context, opts ...grpc.CallOption) (Task_EmitOutputClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Task_serviceDesc.Streams[0], c.cc, "/nomadatc.Task/EmitOutput", opts...)
	if err != nil {
		return nil, err
	}
	x := &taskEmitOutputClient{stream}
	return x, nil
}

type Task_EmitOutputClient interface {
	Send(*OutputData) error
	Recv() (*Actions, error)
	grpc.ClientStream
}

type taskEmitOutputClient struct {
	grpc.ClientStream
}

func (x *taskEmitOutputClient) Send(m *OutputData) error {
	return x.ClientStream.SendMsg(m)
}

func (x *taskEmitOutputClient) Recv() (*Actions, error) {
	m := new(Actions)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *taskClient) ProvideFiles(ctx context.Context, opts ...grpc.CallOption) (Task_ProvideFilesClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Task_serviceDesc.Streams[1], c.cc, "/nomadatc.Task/ProvideFiles", opts...)
	if err != nil {
		return nil, err
	}
	x := &taskProvideFilesClient{stream}
	return x, nil
}

type Task_ProvideFilesClient interface {
	Send(*FileData) error
	Recv() (*FileRequest, error)
	grpc.ClientStream
}

type taskProvideFilesClient struct {
	grpc.ClientStream
}

func (x *taskProvideFilesClient) Send(m *FileData) error {
	return x.ClientStream.SendMsg(m)
}

func (x *taskProvideFilesClient) Recv() (*FileRequest, error) {
	m := new(FileRequest)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for Task service

type TaskServer interface {
	// Sends a greeting
	EmitOutput(Task_EmitOutputServer) error
	ProvideFiles(Task_ProvideFilesServer) error
}

func RegisterTaskServer(s *grpc.Server, srv TaskServer) {
	s.RegisterService(&_Task_serviceDesc, srv)
}

func _Task_EmitOutput_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(TaskServer).EmitOutput(&taskEmitOutputServer{stream})
}

type Task_EmitOutputServer interface {
	Send(*Actions) error
	Recv() (*OutputData, error)
	grpc.ServerStream
}

type taskEmitOutputServer struct {
	grpc.ServerStream
}

func (x *taskEmitOutputServer) Send(m *Actions) error {
	return x.ServerStream.SendMsg(m)
}

func (x *taskEmitOutputServer) Recv() (*OutputData, error) {
	m := new(OutputData)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Task_ProvideFiles_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(TaskServer).ProvideFiles(&taskProvideFilesServer{stream})
}

type Task_ProvideFilesServer interface {
	Send(*FileRequest) error
	Recv() (*FileData, error)
	grpc.ServerStream
}

type taskProvideFilesServer struct {
	grpc.ServerStream
}

func (x *taskProvideFilesServer) Send(m *FileRequest) error {
	return x.ServerStream.SendMsg(m)
}

func (x *taskProvideFilesServer) Recv() (*FileData, error) {
	m := new(FileData)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Task_serviceDesc = grpc.ServiceDesc{
	ServiceName: "nomadatc.Task",
	HandlerType: (*TaskServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "EmitOutput",
			Handler:       _Task_EmitOutput_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "ProvideFiles",
			Handler:       _Task_ProvideFiles_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "log.proto",
}

func (m *OutputData) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *OutputData) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Stream != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintLog(dAtA, i, uint64(m.Stream))
	}
	if m.Stderr {
		dAtA[i] = 0x10
		i++
		if m.Stderr {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i++
	}
	if len(m.Data) > 0 {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintLog(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	if m.Finished {
		dAtA[i] = 0x20
		i++
		if m.Finished {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i++
	}
	if m.FinishedStatus != 0 {
		dAtA[i] = 0x28
		i++
		i = encodeVarintLog(dAtA, i, uint64(m.FinishedStatus))
	}
	return i, nil
}

func (m *Actions) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Actions) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Signal != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintLog(dAtA, i, uint64(m.Signal))
	}
	if len(m.Input) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintLog(dAtA, i, uint64(len(m.Input)))
		i += copy(dAtA[i:], m.Input)
	}
	if m.CloseInput {
		dAtA[i] = 0x18
		i++
		if m.CloseInput {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i++
	}
	return i, nil
}

func (m *FileRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FileRequest) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Path) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintLog(dAtA, i, uint64(len(m.Path)))
		i += copy(dAtA[i:], m.Path)
	}
	if m.Key != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintLog(dAtA, i, uint64(m.Key))
	}
	return i, nil
}

func (m *FileData) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FileData) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Key != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintLog(dAtA, i, uint64(m.Key))
	}
	if len(m.Data) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintLog(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	return i, nil
}

func encodeFixed64Log(dAtA []byte, offset int, v uint64) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	dAtA[offset+4] = uint8(v >> 32)
	dAtA[offset+5] = uint8(v >> 40)
	dAtA[offset+6] = uint8(v >> 48)
	dAtA[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32Log(dAtA []byte, offset int, v uint32) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintLog(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *OutputData) Size() (n int) {
	var l int
	_ = l
	if m.Stream != 0 {
		n += 1 + sovLog(uint64(m.Stream))
	}
	if m.Stderr {
		n += 2
	}
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovLog(uint64(l))
	}
	if m.Finished {
		n += 2
	}
	if m.FinishedStatus != 0 {
		n += 1 + sovLog(uint64(m.FinishedStatus))
	}
	return n
}

func (m *Actions) Size() (n int) {
	var l int
	_ = l
	if m.Signal != 0 {
		n += 1 + sovLog(uint64(m.Signal))
	}
	l = len(m.Input)
	if l > 0 {
		n += 1 + l + sovLog(uint64(l))
	}
	if m.CloseInput {
		n += 2
	}
	return n
}

func (m *FileRequest) Size() (n int) {
	var l int
	_ = l
	l = len(m.Path)
	if l > 0 {
		n += 1 + l + sovLog(uint64(l))
	}
	if m.Key != 0 {
		n += 1 + sovLog(uint64(m.Key))
	}
	return n
}

func (m *FileData) Size() (n int) {
	var l int
	_ = l
	if m.Key != 0 {
		n += 1 + sovLog(uint64(m.Key))
	}
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovLog(uint64(l))
	}
	return n
}

func sovLog(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozLog(x uint64) (n int) {
	return sovLog(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *OutputData) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&OutputData{`,
		`Stream:` + fmt.Sprintf("%v", this.Stream) + `,`,
		`Stderr:` + fmt.Sprintf("%v", this.Stderr) + `,`,
		`Data:` + fmt.Sprintf("%v", this.Data) + `,`,
		`Finished:` + fmt.Sprintf("%v", this.Finished) + `,`,
		`FinishedStatus:` + fmt.Sprintf("%v", this.FinishedStatus) + `,`,
		`}`,
	}, "")
	return s
}
func (this *Actions) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&Actions{`,
		`Signal:` + fmt.Sprintf("%v", this.Signal) + `,`,
		`Input:` + fmt.Sprintf("%v", this.Input) + `,`,
		`CloseInput:` + fmt.Sprintf("%v", this.CloseInput) + `,`,
		`}`,
	}, "")
	return s
}
func (this *FileRequest) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&FileRequest{`,
		`Path:` + fmt.Sprintf("%v", this.Path) + `,`,
		`Key:` + fmt.Sprintf("%v", this.Key) + `,`,
		`}`,
	}, "")
	return s
}
func (this *FileData) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&FileData{`,
		`Key:` + fmt.Sprintf("%v", this.Key) + `,`,
		`Data:` + fmt.Sprintf("%v", this.Data) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringLog(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *OutputData) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowLog
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: OutputData: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: OutputData: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Stream", wireType)
			}
			m.Stream = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLog
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Stream |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Stderr", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLog
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Stderr = bool(v != 0)
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLog
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthLog
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Finished", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLog
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Finished = bool(v != 0)
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field FinishedStatus", wireType)
			}
			m.FinishedStatus = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLog
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.FinishedStatus |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipLog(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthLog
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Actions) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowLog
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Actions: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Actions: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Signal", wireType)
			}
			m.Signal = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLog
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Signal |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Input", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLog
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthLog
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Input = append(m.Input[:0], dAtA[iNdEx:postIndex]...)
			if m.Input == nil {
				m.Input = []byte{}
			}
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CloseInput", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLog
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.CloseInput = bool(v != 0)
		default:
			iNdEx = preIndex
			skippy, err := skipLog(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthLog
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *FileRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowLog
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: FileRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FileRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Path", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLog
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthLog
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Path = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Key", wireType)
			}
			m.Key = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLog
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Key |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipLog(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthLog
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *FileData) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowLog
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: FileData: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FileData: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Key", wireType)
			}
			m.Key = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLog
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Key |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLog
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthLog
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipLog(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthLog
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipLog(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowLog
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowLog
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowLog
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthLog
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowLog
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipLog(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthLog = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowLog   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("log.proto", fileDescriptorLog) }

var fileDescriptorLog = []byte{
	// 380 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x52, 0x3d, 0x8f, 0xd3, 0x40,
	0x10, 0xf5, 0xc4, 0x71, 0x70, 0x26, 0x11, 0x1f, 0xab, 0x80, 0xac, 0x14, 0x8b, 0xe5, 0x06, 0x17,
	0x28, 0x8a, 0x48, 0x49, 0x81, 0x40, 0x80, 0x44, 0x05, 0x5a, 0x28, 0xe8, 0xa2, 0x25, 0x5e, 0x92,
	0x55, 0x1c, 0xdb, 0x78, 0xd7, 0x48, 0x74, 0x14, 0xfc, 0x00, 0x1a, 0xfe, 0x03, 0x3f, 0xe5, 0xca,
	0x94, 0x57, 0x5e, 0x7c, 0xcd, 0x95, 0xf9, 0x09, 0x27, 0xaf, 0xbf, 0x74, 0xdd, 0x7b, 0x6f, 0x66,
	0x67, 0xde, 0x3c, 0x2d, 0x8e, 0xe3, 0x74, 0xbb, 0xc8, 0xf2, 0x54, 0xa7, 0xc4, 0x4d, 0xd2, 0x03,
	0x8f, 0xb8, 0xde, 0x04, 0xff, 0x00, 0xf1, 0x63, 0xa1, 0xb3, 0x42, 0xbf, 0xe5, 0x9a, 0x93, 0x27,
	0x38, 0x52, 0x3a, 0x17, 0xfc, 0xe0, 0x81, 0x0f, 0xa1, 0xcd, 0x1a, 0x56, 0xeb, 0x91, 0xc8, 0x73,
	0x6f, 0xe0, 0x43, 0xe8, 0xb2, 0x86, 0x11, 0x82, 0xc3, 0x88, 0x6b, 0xee, 0xd9, 0x3e, 0x84, 0x53,
	0x66, 0x30, 0x99, 0xa3, 0xfb, 0x5d, 0x26, 0x52, 0xed, 0x44, 0xe4, 0x0d, 0x4d, 0x77, 0xc7, 0xc9,
	0x33, 0x7c, 0xd0, 0xe2, 0xb5, 0xd2, 0x5c, 0x17, 0xca, 0x73, 0x7c, 0x08, 0x1d, 0x76, 0xbf, 0x95,
	0x3f, 0x1b, 0x35, 0xf8, 0x8a, 0xf7, 0x5e, 0x6f, 0xb4, 0x4c, 0x13, 0x65, 0x76, 0xcb, 0x6d, 0xc2,
	0xe3, 0xce, 0x93, 0x61, 0x64, 0x86, 0x8e, 0x4c, 0xb2, 0x42, 0x1b, 0x4b, 0x53, 0x56, 0x13, 0xf2,
	0x14, 0x27, 0x9b, 0x38, 0x55, 0x62, 0x5d, 0xd7, 0x6c, 0x63, 0x00, 0x8d, 0xf4, 0xa1, 0x52, 0x82,
	0x15, 0x4e, 0xde, 0xcb, 0x58, 0x30, 0xf1, 0xa3, 0x10, 0x4a, 0x57, 0x17, 0x64, 0x5c, 0xef, 0xcc,
	0xec, 0x31, 0x33, 0x98, 0x3c, 0x44, 0x7b, 0x2f, 0x7e, 0x99, 0xb9, 0x0e, 0xab, 0x60, 0xb0, 0x44,
	0xb7, 0x7a, 0x64, 0x32, 0x6a, 0xaa, 0xd0, 0x55, 0xbb, 0x14, 0x06, 0x7d, 0x0a, 0x2f, 0xfe, 0x00,
	0x0e, 0xbf, 0x70, 0xb5, 0x27, 0x2f, 0x11, 0xdf, 0x1d, 0xa4, 0xae, 0x43, 0x26, 0xb3, 0x45, 0x1b,
	0xfd, 0xa2, 0x8f, 0x7d, 0xfe, 0xa8, 0x57, 0x9b, 0xab, 0x03, 0x2b, 0x84, 0x25, 0x90, 0x57, 0x38,
	0xfd, 0x94, 0xa7, 0x3f, 0x65, 0x24, 0xaa, 0xf5, 0x8a, 0x90, 0xbe, 0xb1, 0xf5, 0x33, 0x7f, 0x7c,
	0x57, 0x6b, 0x0e, 0xab, 0x07, 0xbc, 0x79, 0x7e, 0x3c, 0x51, 0xeb, 0xf2, 0x44, 0xad, 0xf3, 0x89,
	0xc2, 0xef, 0x92, 0xc2, 0xff, 0x92, 0xc2, 0x45, 0x49, 0xe1, 0x58, 0x52, 0xb8, 0x2a, 0x29, 0xdc,
	0x94, 0xd4, 0x3a, 0x97, 0x14, 0xfe, 0x5e, 0x53, 0xeb, 0xdb, 0xc8, 0x7c, 0x8f, 0xd5, 0x6d, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x92, 0x65, 0xc6, 0xcb, 0x2b, 0x02, 0x00, 0x00,
}
