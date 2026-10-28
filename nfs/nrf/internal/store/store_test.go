package store

import "testing"

func TestPutGetDelete(t *testing.T) {
	s := New()
	p := Profile{"nfType": "AMF", "nfStatus": "REGISTERED"}
	if s.Put("id-1", p) {
		t.Fatal("first put should not exist")
	}
	got, ok := s.Get("id-1")
	if !ok || got["nfType"] != "AMF" || got["nfInstanceId"] != "id-1" {
		t.Fatalf("get mismatch: %+v ok=%v", got, ok)
	}
	if !s.Put("id-1", p) {
		t.Fatal("second put should exist")
	}
	if !s.Delete("id-1") {
		t.Fatal("delete existing")
	}
	if s.Delete("id-1") {
		t.Fatal("delete missing")
	}
	if _, ok := s.Get("id-1"); ok {
		t.Fatal("get after delete")
	}
}

func TestDiscoverFilters(t *testing.T) {
	s := New()
	s.Put("amf", Profile{"nfType": "AMF", "nfStatus": "REGISTERED"})
	s.Put("smf", Profile{
		"nfType":         "SMF",
		"nfStatus":       "REGISTERED",
		"allowedNfTypes": []any{"AMF"},
	})
	s.Put("pcf", Profile{
		"nfType":         "PCF",
		"nfStatus":       "REGISTERED",
		"allowedNfTypes": []any{"SMF"},
	})

	amfs := s.Discover("AMF", "SMF")
	if len(amfs) != 1 {
		t.Fatalf("discover AMF: %d", len(amfs))
	}
	smfs := s.Discover("SMF", "AMF")
	if len(smfs) != 1 {
		t.Fatalf("discover SMF: %d", len(smfs))
	}
	pcfs := s.Discover("PCF", "AMF")
	if len(pcfs) != 0 {
		t.Fatalf("PCF should reject AMF requester: %d", len(pcfs))
	}
	if n := len(s.List("SMF")); n != 1 {
		t.Fatalf("list SMF: %d", n)
	}
	if s.Count() != 3 {
		t.Fatalf("count=%d", s.Count())
	}
}

func TestGetReturnsCopy(t *testing.T) {
	s := New()
	s.Put("id", Profile{"nfType": "AMF", "nfStatus": "REGISTERED"})
	got, _ := s.Get("id")
	got["nfType"] = "SMF"
	again, _ := s.Get("id")
	if again["nfType"] != "AMF" {
		t.Fatal("store mutated through returned map")
	}
}
