package properties_test

import (
	"testing"

	"donarium/server/internal/properties"
)

func TestParseClassification(t *testing.T) {
	cases := []struct {
		in   string
		want properties.Classification
		ok   bool
	}{
		{"house", properties.ClassificationHouse, true},
		{"House", properties.ClassificationHouse, true},
		{"HOUSE", properties.ClassificationHouse, true},
		{"apartment", properties.ClassificationApartment, true},
		{"condominium", properties.ClassificationCondominium, true},
		{"multi_unit", properties.ClassificationMultiUnit, true},
		{"multi-unit", properties.ClassificationMultiUnit, true},
		{"multiunit", properties.ClassificationMultiUnit, true},
		{"commercial", properties.ClassificationCommercial, true},
		{"other", properties.ClassificationOther, true},
		{"castle", "", false},
		{"", "", false},
	}
	for _, c := range cases {
		got, err := properties.ParseClassification(c.in)
		if c.ok && err != nil {
			t.Errorf("ParseClassification(%q): expected ok, got err %v", c.in, err)
		}
		if !c.ok && err == nil {
			t.Errorf("ParseClassification(%q): expected error, got %v", c.in, got)
		}
		if c.ok && got != c.want {
			t.Errorf("ParseClassification(%q): want %q, got %q", c.in, c.want, got)
		}
	}
}

func TestParseRentalCadence(t *testing.T) {
	cases := []struct {
		in   string
		want properties.RentalCadence
		ok   bool
	}{
		{"monthly", properties.CadenceMonthly, true},
		{"Weekly", properties.CadenceWeekly, true},
		{"DAILY", properties.CadenceDaily, true},
		{"annual", properties.CadenceAnnual, true},
		{"biweekly", "", false},
		{"", "", false},
	}
	for _, c := range cases {
		got, err := properties.ParseRentalCadence(c.in)
		if c.ok && err != nil {
			t.Errorf("ParseRentalCadence(%q): expected ok, got err %v", c.in, err)
		}
		if !c.ok && err == nil {
			t.Errorf("ParseRentalCadence(%q): expected error", c.in)
		}
		if c.ok && got != c.want {
			t.Errorf("ParseRentalCadence(%q): want %q, got %q", c.in, c.want, got)
		}
	}
}

func TestValidateDisplayName(t *testing.T) {
	if err := properties.ValidateDisplayName("A"); err == nil {
		t.Error("expected error for short name")
	}
	if err := properties.ValidateDisplayName(""); err == nil {
		t.Error("expected error for empty")
	}
	if err := properties.ValidateDisplayName("Casa Sol"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// 101 chars should fail
	long := string(make([]byte, 101))
	for i := range long {
		long = long[:i] + "a" + long[i+1:]
	}
	// simpler: use 101 a's
	long = ""
	for i := 0; i < 101; i++ {
		long += "a"
	}
	if err := properties.ValidateDisplayName(long); err == nil {
		t.Error("expected error for long name")
	}
}

func TestValidateStandardRent(t *testing.T) {
	if err := properties.ValidateStandardRent(0); err == nil {
		t.Error("expected error for 0")
	}
	if err := properties.ValidateStandardRent(-100); err == nil {
		t.Error("expected error for negative")
	}
	if err := properties.ValidateStandardRent(120000); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAddressValidate(t *testing.T) {
	ok := properties.Address{Street: "123", City: "Madrid", PostalCode: "28001", Country: "ES"}
	if err := ok.Validate(); err != nil {
		t.Errorf("expected ok, got %v", err)
	}
	bad := properties.Address{Street: "", City: "Madrid", PostalCode: "28001", Country: "ES"}
	if err := bad.Validate(); err == nil {
		t.Error("expected error for empty street")
	}
}
