MOV  DPTR, #0x7b16
MOVX A, @DPTR
INC  A
MOVX @DPTR, A

MOV DPTR, #0xf055
RET
