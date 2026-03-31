package bugle

import "testing"

func TestHornLevel_Worse(t *testing.T) {
	tests := []struct {
		a, b HornLevel
		want bool
	}{
		{HornGreen, HornGreen, false},
		{HornYellow, HornGreen, true},
		{HornGreen, HornYellow, false},
		{HornRed, HornYellow, true},
		{HornBlack, HornRed, true},
		{HornBlack, HornBlack, false},
	}
	for _, tt := range tests {
		if got := tt.a.Worse(tt.b); got != tt.want {
			t.Errorf("%s.Worse(%s) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestWorstHorn(t *testing.T) {
	tests := []struct {
		name   string
		levels []HornLevel
		want   HornLevel
	}{
		{"empty", nil, HornGreen},
		{"all green", []HornLevel{HornGreen, HornGreen}, HornGreen},
		{"one yellow", []HornLevel{HornGreen, HornYellow, HornGreen}, HornYellow},
		{"mixed", []HornLevel{HornGreen, HornRed, HornYellow}, HornRed},
		{"black wins", []HornLevel{HornYellow, HornBlack, HornRed}, HornBlack},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WorstHorn(tt.levels...); got != tt.want {
				t.Errorf("WorstHorn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProtocolError(t *testing.T) {
	err := &ProtocolError{Code: ErrCodeNoActiveSession, Message: "session not found"}
	want := "no_active_session: session not found"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestValidStatuses(t *testing.T) {
	for _, s := range []SubmitStatus{StatusOk, StatusBlocked, StatusResolved, StatusCanceled, StatusError} {
		if !ValidStatuses[s] {
			t.Errorf("ValidStatuses[%q] = false, want true", s)
		}
	}
	if ValidStatuses["bogus"] {
		t.Error("ValidStatuses[bogus] = true, want false")
	}
}
