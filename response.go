package ibgo

// import (
// 	"bytes"
// 	"errors"
// 	"fmt"
// 	"time"
// )

// // type Message struct {
// // Code          string
// // ID            string
// // Body          interface{}
// // Error         error
// // bytes         []byte
// // Time          time.Time
// // buf           *bytes.Buffer
// // serverVersion int64
// // version       int64
// // }

// var (
// 	ErrIgnoreResponse error    = errors.New("Ignored message")
// 	ErrReadonly       error    = errors.New("READONLY")
// 	ErrDuplicateReqID error    = errors.New("DUPLICATE")
// 	ErrIOError        error    = errors.New("IOERROR")
// 	ErrMsgTimeout     error    = errors.New("Msg timeout")
// 	MsgTimeoutErr     *Message = &Message{Error: ErrMsgTimeout}
// 	// ResponseReadonly       *Message = &Message{Error: ErrReadonly}
// 	// ResponseDuplicateReqID *Message = &Message{Error: ErrDuplicateReqID}
// 	// ResponseIOError        *Message = &Message{Error: ErrIOError}
// 	// EOF               Response = Response{code: "EOF"}
// )

type TagValue struct {
	Tag   string
	Value string
}

// func RequestSuccess(id string) *Message {
// 	return &Message{ID: id}
// }
// func newMessage(b []byte) *Message {
// 	bytes := make([]byte, len(b))
// 	copy(bytes, b)
// 	return &Message{bytes: bytes, Time: time.Now()}
// }

// type Readable interface {
// 	// return id skip versions
// 	id() string
// 	read()
// }

// type ErrMsg struct {
// 	ErrorCode   int64
// 	ErrorString string
// }

// func (m *ErrMsg) read(msg *Message) {
// 	m.ErrorCode = msg.readInt()
// 	m.ErrorString = msg.readString()
// }

// func (m *ErrMsg) toError() error {
// 	return fmt.Errorf("Request Failed with ErrorCode %v: %s", m.ErrorCode, m.ErrorString)
// }

// type AcctValue struct {
// 	Message
// 	Key         string
// 	Val         string
// 	Currency    string
// 	AccountName string
// }

// func (resp *AcctValue) read() {
// 	readString(resp.buf)
// 	resp.Key = readString(resp.buf)
// 	resp.Val = readString(resp.buf)
// 	resp.Currency = readString(resp.buf)
// 	resp.AccountName = readString(resp.buf)
// }

// type PortfolioValue struct {
// 	Message
// 	serverVersion int64
// 	version       int64
// }

// func (resp *PortfolioValue) read() {
// 	//TODO
// 	resp.version = readInt(resp.buf)
// }

// type AcctUpdateTime struct {
// 	Message
// }

// func (resp *AcctUpdateTime) read() {}

// type NextValidID struct {
// 	Message
// 	ID int64
// }

// func (resp *NextValidID) read() {
// 	readString(resp.buf)
// 	resp.ID = readInt(resp.buf)
// }

// type ExecutionData struct {
// 	Message
// 	version int64
// }

// func (resp *ExecutionData) read() {}

// type UpdateNewsBulletin struct {
// 	Message
// 	MsgID      int64
// 	MsgType    int64
// 	NewMessage string
// 	OriginExch string
// }

// func (resp *UpdateNewsBulletin) id() string {
// 	readString(resp.buf)
// 	return ""
// }

// func (resp *UpdateNewsBulletin) read() {}

// type ManagedAccounts struct {
// 	Message
// 	AccountsList []string
// }

// func (resp *ManagedAccounts) id() string { return "" }

// func (resp *ManagedAccounts) read() {
// 	readString(resp.buf)
// 	resp.AccountsList = readStringList(resp.buf, ",")
// }

// type ReceiveFA struct {
// 	Message
// }

// func (resp *ReceiveFA) id() string { return "" }

// func (resp *ReceiveFA) read() {}

// type CurrentTime struct {
// 	Message
// 	Time string
// }

// func (resp *CurrentTime) id() string {
// 	readString(resp.buf)
// 	return "*" + inCURRENTTIME
// }

// func (resp *CurrentTime) read() {
// 	resp.Time = readString(resp.buf)
// 	fmt.Println("in read", resp.Time)
// }

// func (resp *Message) open(serverVersion int64) {
// 	defer func() {
// 		r := recover()
// 		if r != nil {
// 			resp.Error = fmt.Errorf("Panic during open %v", r)
// 			resp.ID = "-1"
// 		}
// 	}()
// 	resp.buf = bytes.NewBuffer(resp.bytes)
// 	resp.Code = readString(resp.buf)
// 	resp.serverVersion = serverVersion
// 	switch resp.Code {
// 	case inTICKPRICE:
// 		readString(resp.buf)
// 		resp.ID = readString(resp.buf)
// 	case inTICKSIZE:
// 		resp.ID = "IGNORE"
// 	case inORDERSTATUS:
// 		if serverVersion < vMINSERVERVERMARKETCAPPRICE {
// 			readString(resp.buf)
// 			resp.ID = readString(resp.buf)
// 		} else {
// 			resp.ID = readString(resp.buf)
// 		}
// 	case inERRMSG:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inOPENORDER:
// 		if serverVersion < vMINSERVERVERORDERCONTAINER {
// 			resp.version = resp.readInt()
// 		} else {
// 			resp.version = serverVersion
// 		}
// 		resp.ID = "*" + inOPENORDER
// 	case inACCTVALUE:
// 		resp.readString()
// 		resp.ID = "*" + inACCTVALUE
// 	case inPORTFOLIOVALUE:
// 		resp.version = readInt(resp.buf)
// 		resp.ID = "*" + inPORTFOLIOVALUE
// 	case inACCTUPDATETIME:
// 		resp.readString()
// 		resp.ID = "*" + inACCTUPDATETIME
// 	case inNEXTVALIDID:
// 		resp.readString()
// 		resp.ID = "*" + inNEXTVALIDID
// 	case inCONTRACTDATA:
// 		resp.version = resp.readInt()
// 		resp.ID = resp.readString()
// 	case inEXECUTIONDATA:
// 		resp.version = serverVersion
// 		if serverVersion < vMINSERVERVERLASTLIQUIDITY {
// 			resp.version = resp.readInt()
// 		}
// 		resp.ID = resp.readString()
// 	case inMARKETDEPTH:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inMARKETDEPTHL2:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inNEWSBULLETINS:
// 		resp.readString()
// 		resp.ID = "*" + inNEWSBULLETINS
// 	case inMANAGEDACCTS:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inRECEIVEFA:
// 		resp.readString()
// 		resp.ID = "*" + inRECEIVEFA
// 	case inHISTORICALDATA:
// 		resp.ID = resp.readString()
// 	case inBONDCONTRACTDATA:
// 		resp.version = resp.readInt()
// 		resp.ID = resp.readString()
// 	case inSCANNERPARAMETERS:
// 		resp.readString()
// 		resp.ID = "*" + inSCANNERPARAMETERS
// 	case inSCANNERDATA:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inTICKOPTIONCOMPUTATION:
// 		resp.version = resp.readInt()
// 		resp.ID = resp.readString()
// 	case inTICKGENERIC:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inTICKSTRING:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inTICKEFP:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inCURRENTTIME:
// 		resp.readString()
// 		resp.ID = "*" + inCURRENTTIME
// 	case inREALTIMEBARS:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inFUNDAMENTALDATA:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inCONTRACTDATAEND:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inOPENORDEREND:
// 		resp.readString()
// 		resp.ID = "*" + inOPENORDEREND
// 	case inACCTDOWNLOADEND:
// 		resp.readString()
// 		resp.ID = "*" + inACCTDOWNLOADEND
// 	case inEXECUTIONDATAEND:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inDELTANEUTRALVALIDATION:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inTICKSNAPSHOTEND:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inMARKETDATATYPE:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inCOMMISSIONREPORT:
// 		resp.readString()
// 		resp.ID = "*" + inCOMMISSIONREPORT
// 	case inPOSITIONDATA:
// 		resp.version = resp.readInt()
// 		resp.ID = "*" + inPOSITIONDATA
// 	case inPOSITIONEND:
// 		resp.readString()
// 		resp.ID = "*" + inPOSITIONEND
// 	case inACCOUNTSUMMARY:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inACCOUNTSUMMARYEND:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inVERIFYMESSAGEAPI:
// 		resp.ID = "IGNORE"
// 	case inVERIFYCOMPLETED:
// 		resp.ID = "*" + inVERIFYCOMPLETED
// 	case inDISPLAYGROUPLIST:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inDISPLAYGROUPUPDATED:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inVERIFYANDAUTHMESSAGEAPI:
// 		resp.readString()
// 		resp.ID = "*" + inVERIFYMESSAGEAPI
// 	case inVERIFYANDAUTHCOMPLETED:
// 		resp.readString()
// 		resp.ID = "*" + inVERIFYANDAUTHCOMPLETED
// 	case inPOSITIONMULTI:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inPOSITIONMULTIEND:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inACCOUNTUPDATEMULTI:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inACCOUNTUPDATEMULTIEND:
// 		resp.readString()
// 		resp.ID = resp.readString()
// 	case inSECURITYDEFINITIONOPTIONPARAMETER:
// 		resp.ID = resp.readString()
// 	case inSECURITYDEFINITIONOPTIONPARAMETEREND:
// 		resp.ID = resp.readString()
// 	case inSOFTDOLLARTIERS:
// 		resp.ID = resp.readString()
// 	case inFAMILYCODES:
// 		resp.ID = "*" + inFAMILYCODES
// 	case inSYMBOLSAMPLES:
// 		resp.ID = resp.readString()
// 	case inMKTDEPTHEXCHANGES:
// 		resp.ID = "*" + inMKTDEPTHEXCHANGES
// 	case inTICKREQPARAMS:
// 		resp.ID = resp.readString()
// 	case inSMARTCOMPONENTS:
// 		resp.ID = resp.readString()
// 	case inNEWSARTICLE:
// 		resp.ID = resp.readString()
// 	case inTICKNEWS:
// 		resp.ID = resp.readString()
// 	case inNEWSPROVIDERS:
// 		resp.ID = "*" + inNEWSPROVIDERS
// 	case inHISTORICALNEWS:
// 		resp.ID = resp.readString()
// 	case inHISTORICALNEWSEND:
// 		resp.ID = resp.readString()
// 	case inHEADTIMESTAMP:
// 		resp.ID = resp.readString()
// 	case inHISTOGRAMDATA:
// 		resp.ID = resp.readString()
// 	case inHISTORICALDATAUPDATE:
// 		resp.ID = resp.readString()
// 	case inREROUTEMKTDATAREQ:
// 		resp.ID = resp.readString()
// 	case inREROUTEMKTDEPTHREQ:
// 		resp.ID = resp.readString()
// 	case inMARKETRULE:
// 		// TODO special id handle
// 		resp.ID = "*" + inMARKETRULE
// 	case inPNL:
// 		resp.ID = resp.readString()
// 	case inPNLSINGLE:
// 		resp.ID = resp.readString()
// 	case inHISTORICALTICKS:
// 		resp.ID = resp.readString()
// 	case inHISTORICALTICKSBIDASK:
// 		resp.ID = resp.readString()
// 	case inHISTORICALTICKSLAST:
// 		resp.ID = resp.readString()
// 	case inTICKBYTICK:
// 		resp.ID = resp.readString()
// 	case inORDERBOUND:
// 		resp.ID = resp.readString()
// 	case inCOMPLETEDORDER:
// 		resp.ID = "*" + inCOMPLETEDORDER
// 	case inCOMPLETEDORDERSEND:
// 		resp.ID = "*" + inCOMPLETEDORDERSEND
// 	case inREPLACEFAEND:
// 		resp.ID = "*" + inREPLACEFAEND
// 	default:
// 		resp.ID = "-1"
// 	}
// }
