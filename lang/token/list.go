// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package token

// Flags are the low 16 bits of an ID. For example, they signify whether a
// token is an operator, an identifier, etc.
//
// The flags are not exclusive. For example, the "+" token is both a unary and
// binary operator. It is also associative: (a + b) + c equals a + (b + c).
//
// A valid token will have non-zero flags. If none of the other flags apply,
// the FlagsOther bit will be set.
type Flags uint32

const (
	FlagsOther             = Flags(0x0001)
	FlagsUnaryOp           = Flags(0x0002)
	FlagsBinaryOp          = Flags(0x0004)
	FlagsAssociativeOp     = Flags(0x0008)
	FlagsLiteral           = Flags(0x0010)
	FlagsNumLiteral        = Flags(0x0020)
	FlagsStrLiteral        = Flags(0x0040)
	FlagsIdent             = Flags(0x0080)
	FlagsOpen              = Flags(0x0100)
	FlagsClose             = Flags(0x0200)
	FlagsTightLeft         = Flags(0x0400)
	FlagsTightRight        = Flags(0x0800)
	FlagsAssign            = Flags(0x1000)
	FlagsImplicitSemicolon = Flags(0x2000)
	FlagsNumType           = Flags(0x4000)
	flagsUnused            = Flags(0x8000)
)

// Key is the high 16 bits of an ID. It is the map key for an IDMap.
type Key uint32

const (
	FlagsBits = 16
	FlagsMask = 1<<FlagsBits - 1
	KeyBits   = 32 - FlagsBits
	KeyShift  = FlagsBits
	maxKey    = 1<<KeyBits - 1
)

// ID combines a Key and Flags.
type ID uint32

// String returns a string form of x.
func (x ID) String(m *IDMap) string { return m.ByID(x) }

func (x ID) AmbiguousForm() ID   { return ambiguousForms[0xFF&(x>>KeyShift)] }
func (x ID) UnaryForm() ID       { return unaryForms[0xFF&(x>>KeyShift)] }
func (x ID) BinaryForm() ID      { return binaryForms[0xFF&(x>>KeyShift)] }
func (x ID) AssociativeForm() ID { return associativeForms[0xFF&(x>>KeyShift)] }

func (x ID) Key() Key     { return Key(x >> KeyShift) }
func (x ID) Flags() Flags { return Flags(x & FlagsMask) }

func (x ID) IsBuiltIn() bool { return Key(x>>KeyShift) < nBuiltInKeys }

func (x ID) IsUnaryOp() bool           { return Flags(x)&FlagsUnaryOp != 0 }
func (x ID) IsBinaryOp() bool          { return Flags(x)&FlagsBinaryOp != 0 }
func (x ID) IsAssociativeOp() bool     { return Flags(x)&FlagsAssociativeOp != 0 }
func (x ID) IsLiteral() bool           { return Flags(x)&FlagsLiteral != 0 }
func (x ID) IsNumLiteral() bool        { return Flags(x)&FlagsNumLiteral != 0 }
func (x ID) IsStrLiteral() bool        { return Flags(x)&FlagsStrLiteral != 0 }
func (x ID) IsIdent() bool             { return Flags(x)&FlagsIdent != 0 }
func (x ID) IsOpen() bool              { return Flags(x)&FlagsOpen != 0 }
func (x ID) IsClose() bool             { return Flags(x)&FlagsClose != 0 }
func (x ID) IsTightLeft() bool         { return Flags(x)&FlagsTightLeft != 0 }
func (x ID) IsTightRight() bool        { return Flags(x)&FlagsTightRight != 0 }
func (x ID) IsAssign() bool            { return Flags(x)&FlagsAssign != 0 }
func (x ID) IsImplicitSemicolon() bool { return Flags(x)&FlagsImplicitSemicolon != 0 }
func (x ID) IsNumType() bool           { return Flags(x)&FlagsNumType != 0 }

// QID is a qualified ID, such as "foo.bar". QID[0] is "foo"'s ID and QID[1] is
// "bar"'s. QID[0] may be 0 for a plain "bar".
type QID [2]ID

// String returns a string form of x.
func (x QID) String(m *IDMap) string {
	if x[0] == 0 {
		return m.ByID(x[1])
	}
	return m.ByID(x[0]) + "." + m.ByID(x[1])
}

// Token combines an ID and the line number it was seen.
type Token struct {
	ID   ID
	Line uint32
}

func (t Token) Key() Key     { return Key(t.ID >> KeyShift) }
func (t Token) Flags() Flags { return Flags(t.ID & FlagsMask) }

func (t Token) IsBuiltIn() bool { return Key(t.ID>>KeyShift) < nBuiltInKeys }

func (t Token) IsUnaryOp() bool           { return Flags(t.ID)&FlagsUnaryOp != 0 }
func (t Token) IsBinaryOp() bool          { return Flags(t.ID)&FlagsBinaryOp != 0 }
func (t Token) IsAssociativeOp() bool     { return Flags(t.ID)&FlagsAssociativeOp != 0 }
func (t Token) IsLiteral() bool           { return Flags(t.ID)&FlagsLiteral != 0 }
func (t Token) IsNumLiteral() bool        { return Flags(t.ID)&FlagsNumLiteral != 0 }
func (t Token) IsStrLiteral() bool        { return Flags(t.ID)&FlagsStrLiteral != 0 }
func (t Token) IsIdent() bool             { return Flags(t.ID)&FlagsIdent != 0 }
func (t Token) IsOpen() bool              { return Flags(t.ID)&FlagsOpen != 0 }
func (t Token) IsClose() bool             { return Flags(t.ID)&FlagsClose != 0 }
func (t Token) IsTightLeft() bool         { return Flags(t.ID)&FlagsTightLeft != 0 }
func (t Token) IsTightRight() bool        { return Flags(t.ID)&FlagsTightRight != 0 }
func (t Token) IsAssign() bool            { return Flags(t.ID)&FlagsAssign != 0 }
func (t Token) IsImplicitSemicolon() bool { return Flags(t.ID)&FlagsImplicitSemicolon != 0 }
func (t Token) IsNumType() bool           { return Flags(t.ID)&FlagsNumType != 0 }

// nBuiltInKeys is the number of built-in Keys. The packing is:
//  - Zero is invalid.
//  - [0x01, 0x2F] are squiggly punctuation, such as "(", ")" and ";".
//  - [0x30, 0x3F] are squiggly assignments, such as "=" and "+=".
//  - [0x40, 0x5F] are operators, such as "+", "==" and "not".
//  - [0x60, 0x8F] are keywords, such as "if" and "return".
//  - [0x90, 0x9F] are literals, such as "false" and "true".
//  - [0xA0, 0xBF] are identifiers, such as "bool" and "u32".
//  - [0xC0, 0xFF] are disambiguation forms, e.g. unary "+" vs binary "+".
//
// "Squiggly" means a sequence of non-alpha-numeric characters, such as "+" and
// "&=". Their Keys range in [0x01, 0x7F].
const nBuiltInKeys = Key(256)

const (
	KeyInvalid = Key(0)

	KeyOpenParen    = Key(IDOpenParen >> KeyShift)
	KeyCloseParen   = Key(IDCloseParen >> KeyShift)
	KeyOpenBracket  = Key(IDOpenBracket >> KeyShift)
	KeyCloseBracket = Key(IDCloseBracket >> KeyShift)
	KeyOpenCurly    = Key(IDOpenCurly >> KeyShift)
	KeyCloseCurly   = Key(IDCloseCurly >> KeyShift)

	KeyDot       = Key(IDDot >> KeyShift)
	KeyDotDot    = Key(IDDotDot >> KeyShift)
	KeyComma     = Key(IDComma >> KeyShift)
	KeyExclam    = Key(IDExclam >> KeyShift)
	KeyQuestion  = Key(IDQuestion >> KeyShift)
	KeyColon     = Key(IDColon >> KeyShift)
	KeySemicolon = Key(IDSemicolon >> KeyShift)

	KeyEq       = Key(IDEq >> KeyShift)
	KeyPlusEq   = Key(IDPlusEq >> KeyShift)
	KeyMinusEq  = Key(IDMinusEq >> KeyShift)
	KeyStarEq   = Key(IDStarEq >> KeyShift)
	KeySlashEq  = Key(IDSlashEq >> KeyShift)
	KeyShiftLEq = Key(IDShiftLEq >> KeyShift)
	KeyShiftREq = Key(IDShiftREq >> KeyShift)
	KeyAmpEq    = Key(IDAmpEq >> KeyShift)
	KeyAmpHatEq = Key(IDAmpHatEq >> KeyShift)
	KeyPipeEq   = Key(IDPipeEq >> KeyShift)
	KeyHatEq    = Key(IDHatEq >> KeyShift)

	KeyPlus   = Key(IDPlus >> KeyShift)
	KeyMinus  = Key(IDMinus >> KeyShift)
	KeyStar   = Key(IDStar >> KeyShift)
	KeySlash  = Key(IDSlash >> KeyShift)
	KeyShiftL = Key(IDShiftL >> KeyShift)
	KeyShiftR = Key(IDShiftR >> KeyShift)
	KeyAmp    = Key(IDAmp >> KeyShift)
	KeyAmpHat = Key(IDAmpHat >> KeyShift)
	KeyPipe   = Key(IDPipe >> KeyShift)
	KeyHat    = Key(IDHat >> KeyShift)

	KeyNotEq       = Key(IDNotEq >> KeyShift)
	KeyLessThan    = Key(IDLessThan >> KeyShift)
	KeyLessEq      = Key(IDLessEq >> KeyShift)
	KeyEqEq        = Key(IDEqEq >> KeyShift)
	KeyGreaterEq   = Key(IDGreaterEq >> KeyShift)
	KeyGreaterThan = Key(IDGreaterThan >> KeyShift)

	// TODO: sort these by name, when the list has stabilized.
	KeyAnd = Key(IDAnd >> KeyShift)
	KeyOr  = Key(IDOr >> KeyShift)
	KeyNot = Key(IDNot >> KeyShift)
	KeyAs  = Key(IDAs >> KeyShift)

	// TODO: sort these by name, when the list has stabilized.
	KeyFunc     = Key(IDFunc >> KeyShift)
	KeyPtr      = Key(IDPtr >> KeyShift)
	KeyAssert   = Key(IDAssert >> KeyShift)
	KeyWhile    = Key(IDWhile >> KeyShift)
	KeyIf       = Key(IDIf >> KeyShift)
	KeyElse     = Key(IDElse >> KeyShift)
	KeyReturn   = Key(IDReturn >> KeyShift)
	KeyBreak    = Key(IDBreak >> KeyShift)
	KeyContinue = Key(IDContinue >> KeyShift)
	KeyStruct   = Key(IDStruct >> KeyShift)
	KeyUse      = Key(IDUse >> KeyShift)
	KeyVar      = Key(IDVar >> KeyShift)
	KeyIdeal    = Key(IDIdeal >> KeyShift)
	KeyPre      = Key(IDPre >> KeyShift)
	KeyPost     = Key(IDPost >> KeyShift)
	KeyVia      = Key(IDVia >> KeyShift)

	KeyFalse = Key(IDFalse >> KeyShift)
	KeyTrue  = Key(IDTrue >> KeyShift)

	KeyI8    = Key(IDI8 >> KeyShift)
	KeyI16   = Key(IDI16 >> KeyShift)
	KeyI32   = Key(IDI32 >> KeyShift)
	KeyI64   = Key(IDI64 >> KeyShift)
	KeyU8    = Key(IDU8 >> KeyShift)
	KeyU16   = Key(IDU16 >> KeyShift)
	KeyU32   = Key(IDU32 >> KeyShift)
	KeyU64   = Key(IDU64 >> KeyShift)
	KeyUsize = Key(IDUsize >> KeyShift)
	KeyBool  = Key(IDBool >> KeyShift)
	KeyBuf1  = Key(IDBuf1 >> KeyShift)
	KeyBuf2  = Key(IDBuf2 >> KeyShift)

	KeyUnderscore = Key(IDUnderscore >> KeyShift)
	KeyThis       = Key(IDThis >> KeyShift)
	KeyIn         = Key(IDIn >> KeyShift)
	KeyOut        = Key(IDOut >> KeyShift)

	KeyXUnaryPlus  = Key(IDXUnaryPlus >> KeyShift)
	KeyXUnaryMinus = Key(IDXUnaryMinus >> KeyShift)
	KeyXUnaryNot   = Key(IDXUnaryNot >> KeyShift)

	KeyXBinaryPlus        = Key(IDXBinaryPlus >> KeyShift)
	KeyXBinaryMinus       = Key(IDXBinaryMinus >> KeyShift)
	KeyXBinaryStar        = Key(IDXBinaryStar >> KeyShift)
	KeyXBinarySlash       = Key(IDXBinarySlash >> KeyShift)
	KeyXBinaryShiftL      = Key(IDXBinaryShiftL >> KeyShift)
	KeyXBinaryShiftR      = Key(IDXBinaryShiftR >> KeyShift)
	KeyXBinaryAmp         = Key(IDXBinaryAmp >> KeyShift)
	KeyXBinaryAmpHat      = Key(IDXBinaryAmpHat >> KeyShift)
	KeyXBinaryPipe        = Key(IDXBinaryPipe >> KeyShift)
	KeyXBinaryHat         = Key(IDXBinaryHat >> KeyShift)
	KeyXBinaryNotEq       = Key(IDXBinaryNotEq >> KeyShift)
	KeyXBinaryLessThan    = Key(IDXBinaryLessThan >> KeyShift)
	KeyXBinaryLessEq      = Key(IDXBinaryLessEq >> KeyShift)
	KeyXBinaryEqEq        = Key(IDXBinaryEqEq >> KeyShift)
	KeyXBinaryGreaterEq   = Key(IDXBinaryGreaterEq >> KeyShift)
	KeyXBinaryGreaterThan = Key(IDXBinaryGreaterThan >> KeyShift)
	KeyXBinaryAnd         = Key(IDXBinaryAnd >> KeyShift)
	KeyXBinaryOr          = Key(IDXBinaryOr >> KeyShift)
	KeyXBinaryAs          = Key(IDXBinaryAs >> KeyShift)

	KeyXAssociativePlus = Key(IDXAssociativePlus >> KeyShift)
	KeyXAssociativeStar = Key(IDXAssociativeStar >> KeyShift)
	KeyXAssociativeAmp  = Key(IDXAssociativeAmp >> KeyShift)
	KeyXAssociativePipe = Key(IDXAssociativePipe >> KeyShift)
	KeyXAssociativeHat  = Key(IDXAssociativeHat >> KeyShift)
	KeyXAssociativeAnd  = Key(IDXAssociativeAnd >> KeyShift)
	KeyXAssociativeOr   = Key(IDXAssociativeOr >> KeyShift)
)

const (
	IDInvalid = ID(0)

	IDOpenParen    = ID(0x10<<KeyShift | FlagsOpen | FlagsTightRight)
	IDCloseParen   = ID(0x11<<KeyShift | FlagsClose | FlagsTightLeft | FlagsImplicitSemicolon)
	IDOpenBracket  = ID(0x12<<KeyShift | FlagsOpen | FlagsTightLeft | FlagsTightRight)
	IDCloseBracket = ID(0x13<<KeyShift | FlagsClose | FlagsTightLeft | FlagsImplicitSemicolon)
	IDOpenCurly    = ID(0x14<<KeyShift | FlagsOpen)
	IDCloseCurly   = ID(0x15<<KeyShift | FlagsClose | FlagsImplicitSemicolon)

	IDDot       = ID(0x20<<KeyShift | FlagsTightLeft | FlagsTightRight)
	IDDotDot    = ID(0x21<<KeyShift | FlagsTightLeft | FlagsTightRight)
	IDComma     = ID(0x22<<KeyShift | FlagsTightLeft)
	IDExclam    = ID(0x23<<KeyShift | FlagsTightLeft | FlagsTightRight)
	IDQuestion  = ID(0x24<<KeyShift | FlagsTightLeft | FlagsTightRight)
	IDColon     = ID(0x25<<KeyShift | FlagsTightLeft | FlagsTightRight)
	IDSemicolon = ID(0x26<<KeyShift | FlagsTightLeft)

	IDEq       = ID(0x30<<KeyShift | FlagsAssign)
	IDPlusEq   = ID(0x31<<KeyShift | FlagsAssign)
	IDMinusEq  = ID(0x32<<KeyShift | FlagsAssign)
	IDStarEq   = ID(0x33<<KeyShift | FlagsAssign)
	IDSlashEq  = ID(0x34<<KeyShift | FlagsAssign)
	IDShiftLEq = ID(0x35<<KeyShift | FlagsAssign)
	IDShiftREq = ID(0x36<<KeyShift | FlagsAssign)
	IDAmpEq    = ID(0x37<<KeyShift | FlagsAssign)
	IDAmpHatEq = ID(0x38<<KeyShift | FlagsAssign)
	IDPipeEq   = ID(0x39<<KeyShift | FlagsAssign)
	IDHatEq    = ID(0x3A<<KeyShift | FlagsAssign)

	IDPlus   = ID(0x41<<KeyShift | FlagsBinaryOp | FlagsUnaryOp | FlagsAssociativeOp)
	IDMinus  = ID(0x42<<KeyShift | FlagsBinaryOp | FlagsUnaryOp)
	IDStar   = ID(0x43<<KeyShift | FlagsBinaryOp | FlagsAssociativeOp)
	IDSlash  = ID(0x44<<KeyShift | FlagsBinaryOp)
	IDShiftL = ID(0x45<<KeyShift | FlagsBinaryOp)
	IDShiftR = ID(0x46<<KeyShift | FlagsBinaryOp)
	IDAmp    = ID(0x47<<KeyShift | FlagsBinaryOp | FlagsAssociativeOp)
	IDAmpHat = ID(0x48<<KeyShift | FlagsBinaryOp)
	IDPipe   = ID(0x49<<KeyShift | FlagsBinaryOp | FlagsAssociativeOp)
	IDHat    = ID(0x4A<<KeyShift | FlagsBinaryOp | FlagsAssociativeOp)

	IDNotEq       = ID(0x50<<KeyShift | FlagsBinaryOp)
	IDLessThan    = ID(0x51<<KeyShift | FlagsBinaryOp)
	IDLessEq      = ID(0x52<<KeyShift | FlagsBinaryOp)
	IDEqEq        = ID(0x53<<KeyShift | FlagsBinaryOp)
	IDGreaterEq   = ID(0x54<<KeyShift | FlagsBinaryOp)
	IDGreaterThan = ID(0x55<<KeyShift | FlagsBinaryOp)

	// TODO: sort these by name, when the list has stabilized.
	IDAnd = ID(0x58<<KeyShift | FlagsBinaryOp | FlagsAssociativeOp)
	IDOr  = ID(0x59<<KeyShift | FlagsBinaryOp | FlagsAssociativeOp)
	IDNot = ID(0x5A<<KeyShift | FlagsUnaryOp)
	IDAs  = ID(0x5B<<KeyShift | FlagsBinaryOp)

	// TODO: sort these by name, when the list has stabilized.
	IDFunc     = ID(0x60<<KeyShift | FlagsOther)
	IDPtr      = ID(0x61<<KeyShift | FlagsOther)
	IDAssert   = ID(0x62<<KeyShift | FlagsOther)
	IDWhile    = ID(0x63<<KeyShift | FlagsOther)
	IDIf       = ID(0x64<<KeyShift | FlagsOther)
	IDElse     = ID(0x65<<KeyShift | FlagsOther)
	IDReturn   = ID(0x66<<KeyShift | FlagsOther | FlagsImplicitSemicolon)
	IDBreak    = ID(0x67<<KeyShift | FlagsOther | FlagsImplicitSemicolon)
	IDContinue = ID(0x68<<KeyShift | FlagsOther | FlagsImplicitSemicolon)
	IDStruct   = ID(0x69<<KeyShift | FlagsOther)
	IDUse      = ID(0x6A<<KeyShift | FlagsOther)
	IDVar      = ID(0x6B<<KeyShift | FlagsOther)
	IDIdeal    = ID(0x6C<<KeyShift | FlagsOther)
	IDPre      = ID(0x6D<<KeyShift | FlagsOther)
	IDPost     = ID(0x6E<<KeyShift | FlagsOther)
	IDVia      = ID(0x6F<<KeyShift | FlagsOther)

	IDFalse = ID(0x90<<KeyShift | FlagsLiteral | FlagsImplicitSemicolon)
	IDTrue  = ID(0x91<<KeyShift | FlagsLiteral | FlagsImplicitSemicolon)

	IDI8    = ID(0xA0<<KeyShift | FlagsIdent | FlagsImplicitSemicolon | FlagsNumType)
	IDI16   = ID(0xA1<<KeyShift | FlagsIdent | FlagsImplicitSemicolon | FlagsNumType)
	IDI32   = ID(0xA2<<KeyShift | FlagsIdent | FlagsImplicitSemicolon | FlagsNumType)
	IDI64   = ID(0xA3<<KeyShift | FlagsIdent | FlagsImplicitSemicolon | FlagsNumType)
	IDU8    = ID(0xA4<<KeyShift | FlagsIdent | FlagsImplicitSemicolon | FlagsNumType)
	IDU16   = ID(0xA5<<KeyShift | FlagsIdent | FlagsImplicitSemicolon | FlagsNumType)
	IDU32   = ID(0xA6<<KeyShift | FlagsIdent | FlagsImplicitSemicolon | FlagsNumType)
	IDU64   = ID(0xA7<<KeyShift | FlagsIdent | FlagsImplicitSemicolon | FlagsNumType)
	IDUsize = ID(0xA8<<KeyShift | FlagsIdent | FlagsImplicitSemicolon | FlagsNumType)
	IDBool  = ID(0xA9<<KeyShift | FlagsIdent | FlagsImplicitSemicolon)
	IDBuf1  = ID(0xAA<<KeyShift | FlagsIdent | FlagsImplicitSemicolon)
	IDBuf2  = ID(0xAB<<KeyShift | FlagsIdent | FlagsImplicitSemicolon)

	IDUnderscore = ID(0xB0<<KeyShift | FlagsIdent | FlagsImplicitSemicolon)
	IDThis       = ID(0xB1<<KeyShift | FlagsIdent | FlagsImplicitSemicolon)
	IDIn         = ID(0xB2<<KeyShift | FlagsIdent | FlagsImplicitSemicolon)
	IDOut        = ID(0xB3<<KeyShift | FlagsIdent | FlagsImplicitSemicolon)
)

// The IDXFoo IDs are not returned by the tokenizer. They are used by the
// ast.Node ID-typed fields to disambiguate e.g. unary vs binary plus.
const (
	minXKey = 0xC0
	maxXKey = 0xFF

	IDXUnaryPlus  = ID(0xC0<<KeyShift | FlagsUnaryOp)
	IDXUnaryMinus = ID(0xC1<<KeyShift | FlagsUnaryOp)
	IDXUnaryNot   = ID(0xC2<<KeyShift | FlagsUnaryOp)

	IDXBinaryPlus        = ID(0xD0<<KeyShift | FlagsBinaryOp)
	IDXBinaryMinus       = ID(0xD1<<KeyShift | FlagsBinaryOp)
	IDXBinaryStar        = ID(0xD2<<KeyShift | FlagsBinaryOp)
	IDXBinarySlash       = ID(0xD3<<KeyShift | FlagsBinaryOp)
	IDXBinaryShiftL      = ID(0xD4<<KeyShift | FlagsBinaryOp)
	IDXBinaryShiftR      = ID(0xD5<<KeyShift | FlagsBinaryOp)
	IDXBinaryAmp         = ID(0xD6<<KeyShift | FlagsBinaryOp)
	IDXBinaryAmpHat      = ID(0xD7<<KeyShift | FlagsBinaryOp)
	IDXBinaryPipe        = ID(0xD8<<KeyShift | FlagsBinaryOp)
	IDXBinaryHat         = ID(0xD9<<KeyShift | FlagsBinaryOp)
	IDXBinaryNotEq       = ID(0xDA<<KeyShift | FlagsBinaryOp)
	IDXBinaryLessThan    = ID(0xDB<<KeyShift | FlagsBinaryOp)
	IDXBinaryLessEq      = ID(0xDC<<KeyShift | FlagsBinaryOp)
	IDXBinaryEqEq        = ID(0xDD<<KeyShift | FlagsBinaryOp)
	IDXBinaryGreaterEq   = ID(0xDE<<KeyShift | FlagsBinaryOp)
	IDXBinaryGreaterThan = ID(0xDF<<KeyShift | FlagsBinaryOp)
	IDXBinaryAnd         = ID(0xE0<<KeyShift | FlagsBinaryOp)
	IDXBinaryOr          = ID(0xE1<<KeyShift | FlagsBinaryOp)
	IDXBinaryAs          = ID(0xE2<<KeyShift | FlagsBinaryOp)

	IDXAssociativePlus = ID(0xF0<<KeyShift | FlagsAssociativeOp)
	IDXAssociativeStar = ID(0xF1<<KeyShift | FlagsAssociativeOp)
	IDXAssociativeAmp  = ID(0xF2<<KeyShift | FlagsAssociativeOp)
	IDXAssociativePipe = ID(0xF3<<KeyShift | FlagsAssociativeOp)
	IDXAssociativeHat  = ID(0xF4<<KeyShift | FlagsAssociativeOp)
	IDXAssociativeAnd  = ID(0xF5<<KeyShift | FlagsAssociativeOp)
	IDXAssociativeOr   = ID(0xF6<<KeyShift | FlagsAssociativeOp)
)

var builtInsByKey = [nBuiltInKeys]struct {
	name string
	id   ID
}{
	KeyOpenParen:    {"(", IDOpenParen},
	KeyCloseParen:   {")", IDCloseParen},
	KeyOpenBracket:  {"[", IDOpenBracket},
	KeyCloseBracket: {"]", IDCloseBracket},
	KeyOpenCurly:    {"{", IDOpenCurly},
	KeyCloseCurly:   {"}", IDCloseCurly},

	KeyDot:       {".", IDDot},
	KeyDotDot:    {"..", IDDotDot},
	KeyComma:     {",", IDComma},
	KeyExclam:    {"!", IDExclam},
	KeyQuestion:  {"?", IDQuestion},
	KeyColon:     {":", IDColon},
	KeySemicolon: {";", IDSemicolon},

	KeyEq:       {"=", IDEq},
	KeyPlusEq:   {"+=", IDPlusEq},
	KeyMinusEq:  {"-=", IDMinusEq},
	KeyStarEq:   {"*=", IDStarEq},
	KeySlashEq:  {"/=", IDSlashEq},
	KeyShiftLEq: {"<<=", IDShiftLEq},
	KeyShiftREq: {">>=", IDShiftREq},
	KeyAmpEq:    {"&=", IDAmpEq},
	KeyAmpHatEq: {"&^=", IDAmpHatEq},
	KeyPipeEq:   {"|=", IDPipeEq},
	KeyHatEq:    {"^=", IDHatEq},

	KeyPlus:   {"+", IDPlus},
	KeyMinus:  {"-", IDMinus},
	KeyStar:   {"*", IDStar},
	KeySlash:  {"/", IDSlash},
	KeyShiftL: {"<<", IDShiftL},
	KeyShiftR: {">>", IDShiftR},
	KeyAmp:    {"&", IDAmp},
	KeyAmpHat: {"&^", IDAmpHat},
	KeyPipe:   {"|", IDPipe},
	KeyHat:    {"^", IDHat},

	KeyNotEq:       {"!=", IDNotEq},
	KeyLessThan:    {"<", IDLessThan},
	KeyLessEq:      {"<=", IDLessEq},
	KeyEqEq:        {"==", IDEqEq},
	KeyGreaterEq:   {">=", IDGreaterEq},
	KeyGreaterThan: {">", IDGreaterThan},

	KeyAnd: {"and", IDAnd},
	KeyOr:  {"or", IDOr},
	KeyNot: {"not", IDNot},
	KeyAs:  {"as", IDAs},

	KeyFunc:     {"func", IDFunc},
	KeyPtr:      {"ptr", IDPtr},
	KeyAssert:   {"assert", IDAssert},
	KeyWhile:    {"while", IDWhile},
	KeyIf:       {"if", IDIf},
	KeyElse:     {"else", IDElse},
	KeyReturn:   {"return", IDReturn},
	KeyBreak:    {"break", IDBreak},
	KeyContinue: {"continue", IDContinue},
	KeyStruct:   {"struct", IDStruct},
	KeyUse:      {"use", IDUse},
	KeyVar:      {"var", IDVar},
	KeyIdeal:    {"ideal", IDIdeal},
	KeyPre:      {"pre", IDPre},
	KeyPost:     {"post", IDPost},
	KeyVia:      {"via", IDVia},

	KeyFalse: {"false", IDFalse},
	KeyTrue:  {"true", IDTrue},

	KeyI8:    {"i8", IDI8},
	KeyI16:   {"i16", IDI16},
	KeyI32:   {"i32", IDI32},
	KeyI64:   {"i64", IDI64},
	KeyU8:    {"u8", IDU8},
	KeyU16:   {"u16", IDU16},
	KeyU32:   {"u32", IDU32},
	KeyU64:   {"u64", IDU64},
	KeyUsize: {"usize", IDUsize},
	KeyBool:  {"bool", IDBool},
	KeyBuf1:  {"buf1", IDBuf1},
	KeyBuf2:  {"buf2", IDBuf2},

	KeyUnderscore: {"_", IDUnderscore},
	KeyThis:       {"this", IDThis},
	KeyIn:         {"in", IDIn},
	KeyOut:        {"out", IDOut},
}

var builtInsByName = map[string]ID{}

func init() {
	for _, x := range builtInsByKey {
		if x.name != "" {
			builtInsByName[x.name] = x.id
		}
	}
}

// squiggles are built-in IDs that aren't alpha-numeric.
var squiggles = [256]ID{
	'(': IDOpenParen,
	')': IDCloseParen,
	'[': IDOpenBracket,
	']': IDCloseBracket,
	'{': IDOpenCurly,
	'}': IDCloseCurly,

	',': IDComma,
	'?': IDQuestion,
	':': IDColon,
	';': IDSemicolon,
}

type suffixLexer struct {
	suffix string
	id     ID
}

// lexers lex ambiguous 1-byte squiggles. For example, "&" might be the start
// of "&^" or "&=".
//
// The order of the []suffixLexer elements matters. The first match wins. Since
// we want to lex greedily, longer suffixes should be earlier in the slice.
var lexers = [256][]suffixLexer{
	'.': {
		{".", IDDotDot},
		{"", IDDot},
	},
	'!': {
		{"=", IDNotEq},
		{"", IDExclam},
	},
	'&': {
		{"^=", IDAmpHatEq},
		{"^", IDAmpHat},
		{"=", IDAmpEq},
		{"", IDAmp},
	},
	'|': {
		{"=", IDPipeEq},
		{"", IDPipe},
	},
	'^': {
		{"=", IDHatEq},
		{"", IDHat},
	},
	'+': {
		{"=", IDPlusEq},
		{"", IDPlus},
	},
	'-': {
		{"=", IDMinusEq},
		{"", IDMinus},
	},
	'*': {
		{"=", IDStarEq},
		{"", IDStar},
	},
	'/': {
		{"=", IDSlashEq},
		{"", IDSlash},
	},
	'=': {
		{"=", IDEqEq},
		{"", IDEq},
	},
	'<': {
		{"<=", IDShiftLEq},
		{"<", IDShiftL},
		{"=", IDLessEq},
		{"", IDLessThan},
	},
	'>': {
		{">=", IDShiftREq},
		{">", IDShiftR},
		{"=", IDGreaterEq},
		{"", IDGreaterThan},
	},
}

var ambiguousForms = [256]ID{
	KeyXUnaryPlus:  IDPlus,
	KeyXUnaryMinus: IDMinus,
	KeyXUnaryNot:   IDNot,

	KeyXBinaryPlus:        IDPlus,
	KeyXBinaryMinus:       IDMinus,
	KeyXBinaryStar:        IDStar,
	KeyXBinarySlash:       IDSlash,
	KeyXBinaryShiftL:      IDShiftL,
	KeyXBinaryShiftR:      IDShiftR,
	KeyXBinaryAmp:         IDAmp,
	KeyXBinaryAmpHat:      IDAmpHat,
	KeyXBinaryPipe:        IDPipe,
	KeyXBinaryHat:         IDHat,
	KeyXBinaryNotEq:       IDNotEq,
	KeyXBinaryLessThan:    IDLessThan,
	KeyXBinaryLessEq:      IDLessEq,
	KeyXBinaryEqEq:        IDEqEq,
	KeyXBinaryGreaterEq:   IDGreaterEq,
	KeyXBinaryGreaterThan: IDGreaterThan,
	KeyXBinaryAnd:         IDAnd,
	KeyXBinaryOr:          IDOr,
	KeyXBinaryAs:          IDAs,

	KeyXAssociativePlus: IDPlus,
	KeyXAssociativeStar: IDStar,
	KeyXAssociativeAmp:  IDAmp,
	KeyXAssociativePipe: IDPipe,
	KeyXAssociativeHat:  IDHat,
	KeyXAssociativeAnd:  IDAnd,
	KeyXAssociativeOr:   IDOr,
}

var unaryForms = [256]ID{
	KeyPlus:  IDXUnaryPlus,
	KeyMinus: IDXUnaryMinus,
	KeyNot:   IDXUnaryNot,
}

var binaryForms = [256]ID{
	KeyPlusEq:   IDXBinaryPlus,
	KeyMinusEq:  IDXBinaryMinus,
	KeyStarEq:   IDXBinaryStar,
	KeySlashEq:  IDXBinarySlash,
	KeyShiftLEq: IDXBinaryShiftL,
	KeyShiftREq: IDXBinaryShiftR,
	KeyAmpEq:    IDXBinaryAmp,
	KeyAmpHatEq: IDXBinaryAmpHat,
	KeyPipeEq:   IDXBinaryPipe,
	KeyHatEq:    IDXBinaryHat,

	KeyPlus:        IDXBinaryPlus,
	KeyMinus:       IDXBinaryMinus,
	KeyStar:        IDXBinaryStar,
	KeySlash:       IDXBinarySlash,
	KeyShiftL:      IDXBinaryShiftL,
	KeyShiftR:      IDXBinaryShiftR,
	KeyAmp:         IDXBinaryAmp,
	KeyAmpHat:      IDXBinaryAmpHat,
	KeyPipe:        IDXBinaryPipe,
	KeyHat:         IDXBinaryHat,
	KeyNotEq:       IDXBinaryNotEq,
	KeyLessThan:    IDXBinaryLessThan,
	KeyLessEq:      IDXBinaryLessEq,
	KeyEqEq:        IDXBinaryEqEq,
	KeyGreaterEq:   IDXBinaryGreaterEq,
	KeyGreaterThan: IDXBinaryGreaterThan,
	KeyAnd:         IDXBinaryAnd,
	KeyOr:          IDXBinaryOr,
	KeyAs:          IDXBinaryAs,
}

var associativeForms = [256]ID{
	KeyPlus: IDXAssociativePlus,
	KeyStar: IDXAssociativeStar,
	KeyAmp:  IDXAssociativeAmp,
	KeyPipe: IDXAssociativePipe,
	KeyHat:  IDXAssociativeHat,
	KeyAnd:  IDXAssociativeAnd,
	KeyOr:   IDXAssociativeOr,
}
