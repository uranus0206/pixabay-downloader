package dbmanager

import "testing"

func TestDbCreateAndRemove(t *testing.T) {
	InitWithPath("pixabay.db").CreateTable()

	err := DeleteWithPath("pixabay.db")
	if err != nil {
		t.Fatal(err)
	}
}
