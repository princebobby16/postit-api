package tests

import (
	"github.com/twinj/uuid"
	"gitlab.com/pbobby001/postit-api/pkg"
	"net/http"
	"testing"
)

func TestValidateHeaders(t *testing.T) {
	req, err := http.NewRequest("GET", "/health-check", nil)
	if err != nil {
		t.Fatal(err)
	}

	traceId := uuid.NewV4().String()
	tenantNamespace := "postit"
	req.Header.Add("trace-id", traceId)
	req.Header.Add("tenant-namespace", tenantNamespace)

	headers, err := pkg.ValidateHeaders(req)
	if err != nil {
		t.Fatal(err)
	}
	expect := make(map[string]string)
	expect["trace-id"] = traceId
	expect["tenant-namespace"] = tenantNamespace

	if headers["trace-id"] != expect["trace-id"] || headers["tenant-namespace"] != expect["tenant-namespace"] {
		t.Fail()
	}
}
