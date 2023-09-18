package classfile

import (
	"encoding/binary"
)

type Const struct {
	Tag              byte
	NameIndex        uint16
	ClassIndex       uint16
	NameAndTypeIndex uint16
	StringIndex      uint16
	DescIndex        uint16
	Float            float32
	String           string
}

type ConstPool []Const

func (cp *ConstPool) AddConst(c Const) uint16 {
	*cp = append(*cp, c)
	return uint16(len(*cp))
}

type Attribute struct {
	Name string
	Data []byte
}

type Field struct {
	Flags      uint16
	Name       string
	Descriptor string
	Attributes []Attribute
}

type Class struct {
	constPool  ConstPool
	name       string
	super      string
	flags      uint16
	interfaces []string
	fields     []Field
	methods    []Field
	attributes []Attribute
}

func NewClass(name string, super string) *Class {
	class := Class{
		name:       name,
		super:      super,
		constPool:  make(ConstPool, 0),
		flags:      uint16(0),
		interfaces: make([]string, 0),
		fields:     make([]Field, 0),
		methods:    make([]Field, 0),
		attributes: make([]Attribute, 0),
	}
	return &class
}

func (c *Class) AddMethod(name string, descriptor string, byteCode []byte, maxLocalVariables uint16) {
	codeData := make([]byte, 0)
	codeData = binary.BigEndian.AppendUint16(codeData, uint16(0))
	codeData = binary.BigEndian.AppendUint16(codeData, maxLocalVariables)
	codeData = binary.BigEndian.AppendUint32(codeData, uint32(0))
	codeData = append(codeData, byteCode...)
	codeAttribute := Attribute{Name: "Code", Data: codeData}
	c.methods = append(c.methods, Field{Flags: uint16(0), Name: name, Descriptor: descriptor, Attributes: []Attribute{codeAttribute}})
}

func (c *Class) ConvertToBytes() []byte {
	classfile := make([]byte, 0)
	nameIndex := c.constPool.AddConst(Const{Tag: 0x01, String: c.name})
	//setting the access flags
	classfile = binary.BigEndian.AppendUint16(classfile, uint16(0))
	//setting the class index for this and the super class
	classfile = binary.BigEndian.AppendUint16(classfile, nameIndex)
	nameIndex = c.constPool.AddConst(Const{Tag: 0x01, String: c.super})
	classfile = binary.BigEndian.AppendUint16(classfile, nameIndex)
	//setting the length of the interface and field table
	classfile = binary.BigEndian.AppendUint16(classfile, uint16(0))
	classfile = binary.BigEndian.AppendUint16(classfile, uint16(0))
	classfile = binary.BigEndian.AppendUint16(classfile, uint16(len(c.methods)))
	classfile = append(classfile, c.convertMethodsToBytes()...)
	classfile = binary.BigEndian.AppendUint16(classfile, uint16(0))
	finalClassfile := make([]byte, 0)
	constPoolLen := len(c.constPool) + 1
	constPool := c.convertConstPoolToBytes()
	finalClassfile = binary.BigEndian.AppendUint64(finalClassfile, uint64(0))
	finalClassfile = binary.BigEndian.AppendUint16(finalClassfile, uint16(constPoolLen))
	finalClassfile = append(finalClassfile, constPool...)
	finalClassfile = append(finalClassfile, classfile...)
	return finalClassfile
}

func (c *Class) convertMethodsToBytes() []byte {
	allMethodsAsBytes := make([]byte, 0)
	for _, m := range c.methods {
		methodAsBytes := make([]byte, 0)
		methodAsBytes = binary.BigEndian.AppendUint16(methodAsBytes, m.Flags)
		nameIndex := c.constPool.AddConst(Const{Tag: 0x01, String: m.Name})
		descIndex := c.constPool.AddConst(Const{Tag: 0x01, String: m.Descriptor})
		methodAsBytes = binary.BigEndian.AppendUint16(methodAsBytes, nameIndex)
		methodAsBytes = binary.BigEndian.AppendUint16(methodAsBytes, descIndex)
		methodAsBytes = binary.BigEndian.AppendUint16(methodAsBytes, uint16(len(m.Attributes)))
		for _, a := range m.Attributes {
			attributeAsBytes := make([]byte, 0)
			nameIndex = c.constPool.AddConst(Const{Tag: 0x01, String: a.Name})
			attributeAsBytes = binary.BigEndian.AppendUint16(attributeAsBytes, nameIndex)
			attributeAsBytes = binary.BigEndian.AppendUint32(attributeAsBytes, uint32(len(a.Data)))
			attributeAsBytes = append(attributeAsBytes, a.Data...)
			methodAsBytes = append(methodAsBytes, attributeAsBytes...)
		}
		allMethodsAsBytes = append(allMethodsAsBytes, methodAsBytes...)
	}
	return allMethodsAsBytes
}

func (c *Class) convertConstPoolToBytes() []byte {
	constPoolAsBytes := make([]byte, 0)
	for _, co := range c.constPool {
		constAsBytes := make([]byte, 0)
		constAsBytes = append(constAsBytes, co.Tag)
		switch co.Tag {
		case 0x01:
			valueInBytes := []byte(co.String)
			constAsBytes = binary.BigEndian.AppendUint16(constAsBytes, uint16(len(valueInBytes)))
			constAsBytes = append(constAsBytes, valueInBytes...)
		case 0x07:
			constAsBytes = binary.BigEndian.AppendUint16(constAsBytes, co.NameIndex)
		}
		constPoolAsBytes = append(constPoolAsBytes, constAsBytes...)
	}
	return constPoolAsBytes
}
