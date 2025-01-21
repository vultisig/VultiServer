package relay

import "testing"

func TestEncryptGCM(t *testing.T) {
	encryptionKey := "d6022efdbf1cd27b2feb179341b40a800f4fdda7cdfd91ca630f1f17ee0516f3"
	rawInput := "helloworld"
	encryptedResult, err := encryptWrapper(rawInput, encryptionKey, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("encrypted: %s", encryptedResult)
}
