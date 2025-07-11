package utilsCrypto

import (
	"log"
	"testing"
)

func BenchmarkGetBlake256Hash(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		getBlake256Hash([]byte("0x82738128ddidjaxxx77317881"))
	}
}

func BenchmarkVerifyBlake256Hash(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		response := verifyDataBlake256Hash([]byte("0x82738128ddidjaxxx77317881"), "9497a9f56b670821d474360d92956e287dde8c0f6574e2de96da3e0e9ba5210d")
		if !response {
			log.Fatalf("Error VerifyDataBlake256Hash")
		}
	}
}
func BenchmarkCreatePasswordHashArgon2id(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		createPasswordHashArgon2id(defaultPasswordHashParams, "0x82738128ddidjaxxx77317881")
	}
}

func BenchmarkVerifyPasswordHashArgon2id(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err := comparePasswordAndHashArgon2id("0x82738128ddidjaxxx77317881", "$argon2id$v=19$m=32768,t=3,p=1$xdGP8P1vXIohJRDnwjXzaw$YPyDGg6U9DJluiIXUxzhsonOD+HR6tlL7nktr4l5rzs")
		if err != nil {
			log.Fatalf("Error ComparePasswordAndHash")
		}
	}
}
