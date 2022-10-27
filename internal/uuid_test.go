package internal

import "testing"

func TestVerifyUUID(tt *testing.T) {
	tt.Run("Good UUID 1", func(t *testing.T) {
		if IsUUIDFormat("00112233-4455-6677-8899-aabbccddeeff") == false {
			t.Fail()
		}
	})
	tt.Run("Good UUID 2", func(t *testing.T) {
		if IsUUIDFormat("AABBCCDD-EEFF-0011-2233-445566778899") == false {
			t.Fail()
		}
	})
	tt.Run("Bad UUID 1", func(t *testing.T) {
		if IsUUIDFormat("00112233-4455-6677-8899-aabbccddee") == true {
			t.Fail()
		}
	})
	tt.Run("Bad UUID 2", func(t *testing.T) {
		if IsUUIDFormat("00112233-445506677-8899-aabbccddeeff") == true {
			t.Fail()
		}
	})
	tt.Run("Bad UUID 3", func(t *testing.T) {
		if IsUUIDFormat("00112233-44SS-6677-8899-aabbccddeeff") == true {
			t.Fail()
		}
	})
}
