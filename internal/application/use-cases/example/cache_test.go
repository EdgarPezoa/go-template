package example

import "testing"

func TestCacheUseCasesAreConstructible(t *testing.T) {
	if NewGetCacheUseCase(nil) == nil ||
		NewSetCacheUseCase(nil) == nil ||
		NewGetCacheHashUseCase(nil) == nil ||
		NewSetCacheHashUseCase(nil) == nil {
		t.Fatal("nil use case")
	}
}
