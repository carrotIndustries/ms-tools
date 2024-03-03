#include "mslib.h"
#include <stdio.h>
#include <assert.h>
#include <malloc.h>

typedef struct {
  uint8_t mod;
  uint8_t keycode[6];
  uint8_t button;
  int8_t x;
  int8_t y;
  int8_t v;
  int8_t h;
} i2c_msg_t;

int main(void) {
	assert(sizeof(i2c_msg_t) == 12);
	char name[] = "USBKVM";
	uintptr_t handle = MsHalOpen(name);
	printf("handle=%d\n", handle);
	i2c_msg_t msg = {0};
	void *rdData = 0;
	int retval = MsHalI2CTransfer(handle, 10, &msg, sizeof(msg), 0, &rdData);
	printf("write=%d %p\n", retval, rdData);
	free(rdData);
	
	MsHalClose(handle);
	return 0;
}
