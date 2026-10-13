package dto

import "testing"

func TestPackWebhookFormValidate(t *testing.T) {
	valid := func() PackWebhookForm {
		return PackWebhookForm{
			Enabled: true,
			URL:     " https://cow.lan/api/servers/1/restart ",
			Method:  "post",
			Headers: []WebhookHeaderForm{{Name: "Authorization", Value: "Bearer x"}},
		}
	}
	if f := valid(); f.Validate() != nil || f.Method != "POST" || f.URL != "https://cow.lan/api/servers/1/restart" {
		t.Fatalf("valid form rejected or not normalized: %+v", f)
	}

	cases := map[string]func(*PackWebhookForm){
		"blank url when enabled": func(f *PackWebhookForm) { f.URL = "" },
		"ftp scheme":             func(f *PackWebhookForm) { f.URL = "ftp://x/y" },
		"credentials in url":     func(f *PackWebhookForm) { f.URL = "https://u:p@x/y" },
		"bad method":             func(f *PackWebhookForm) { f.Method = "DELETE" },
		"bad header name":        func(f *PackWebhookForm) { f.Headers[0].Name = "bad name" },
		"forbidden header":       func(f *PackWebhookForm) { f.Headers[0].Name = "Host" },
		"duplicate header": func(f *PackWebhookForm) {
			f.Headers = append(f.Headers, WebhookHeaderForm{Name: "authorization", Value: "y"})
		},
		"newline in value": func(f *PackWebhookForm) { f.Headers[0].Value = "a\r\nX: b" },
	}
	for name, mutate := range cases {
		f := valid()
		mutate(&f)
		if f.Validate() == nil {
			t.Errorf("%s: expected error", name)
		}
	}

	off := PackWebhookForm{}
	if err := off.Validate(); err != nil || off.Method != "POST" {
		t.Errorf("empty disabled form: %v %+v", err, off)
	}
}
