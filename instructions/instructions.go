package instructions

const (
	ILOAD   = 0x15
	ILOAD_0 = 0x1a
	ILOAD_1 = 0x1b
	ILOAD_2 = 0x1c
	ILOAD_3 = 0x1d

	IADD = 0x60
	ISUB = 0x64
	IMUL = 0x68
	IDIV = 0x6c

	ISTORE   = 0x36
	ISTORE_0 = 0x3b
	ISTORE_1 = 0x3c
	ISTORE_2 = 0x3d
	ISTORE_3 = 0x3e

	ICONST_M1 = 0x02
	ICONST_0  = 0x03
	ICONST_1  = 0x04
	ICONST_2  = 0x05
	ICONST_3  = 0x06
	ICONST_4  = 0x07
	ICONST_5  = 0x08

	IRETURN = 0xac
)

var Iconsts map[int32]byte = map[int32]byte{
	-1: ICONST_M1,
	0:  ICONST_0,
	1:  ICONST_1,
	2:  ICONST_2,
	3:  ICONST_3,
	4:  ICONST_4,
	5:  ICONST_5,
}

var Istores map[int32]byte = map[int32]byte{
	0: ISTORE_0,
	1: ISTORE_1,
	2: ISTORE_2,
	3: ISTORE_3,
}

var Iloads map[int32]byte = map[int32]byte{
	0: ILOAD_0,
	1: ILOAD_1,
	2: ILOAD_2,
	3: ILOAD_3,
}
