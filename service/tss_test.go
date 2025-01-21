package service

import "testing"

func TestEncryptionGCM(t *testing.T) {
	encryptionKey := "d6022efdbf1cd27b2feb179341b40a800f4fdda7cdfd91ca630f1f17ee0516f3"
	encryptionResult := "lBVUUrBAYm2R6uiESzrgOaaW0GyiOuf2ki6O18YOEBFnQryTj4s="
	encrypted, err := decryptWrapper(encryptionResult, encryptionKey, true)
	if err != nil {
		t.Fatal(err)
	}
	if encrypted != "helloworld" {
		t.Fatalf("decrypted: %s, expected: %s", encrypted, "helloworld")
	}
}
