package client_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stefanoj3/dirstalk/pkg/common/test"
	"github.com/stefanoj3/dirstalk/pkg/scan/client"
	"github.com/stretchr/testify/assert"
)

func TestWhenRemoteIsTooSlowClientShouldTimeout(t *testing.T) {
	testServer, _ := test.NewServerWithAssertion(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(time.Millisecond * 50)
		}),
	)
	defer testServer.Close()

	c, err := client.NewClientFromConfig(
		10,
		nil,
		"",
		false,
		nil,
		nil,
	)
	assert.NoError(t, err)

	res, err := c.Get(testServer.URL)
	assert.Error(t, err)
	assert.Nil(t, res)

	assert.Contains(t, err.Error(), "Client.Timeout")
}

func TestShouldForwardProvidedCookiesWhenUsingJar(t *testing.T) {
	const (
		serverCookieName  = "server_cookie_name"
		serverCookieValue = "server_cookie_value"
	)

	testServer, serverAssertion := test.NewServerWithAssertion(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.SetCookie(
				w,
				&http.Cookie{
					Name:    serverCookieName,
					Value:   serverCookieValue,
					Expires: time.Now().AddDate(0, 1, 0),
				},
			)
		}),
	)
	defer testServer.Close()

	u, err := url.Parse(testServer.URL)
	assert.NoError(t, err)

	cookies := []*http.Cookie{
		{
			Name:  "a_cookie_name",
			Value: "a_cookie_value",
		},
	}

	c, err := client.NewClientFromConfig(
		100,
		nil,
		"",
		true,
		cookies,
		u,
	)
	assert.NoError(t, err)

	res, err := c.Get(testServer.URL)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assert.Equal(t, 1, serverAssertion.Len())

	serverAssertion.At(0, func(r http.Request) {
		assert.Equal(t, 1, len(r.Cookies()))

		assert.Equal(t, r.Cookies()[0].Name, cookies[0].Name)
		assert.Equal(t, r.Cookies()[0].Value, cookies[0].Value)
		assert.Equal(t, r.Cookies()[0].Expires, cookies[0].Expires)
	})

	res, err = c.Get(testServer.URL)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assert.Equal(t, 2, serverAssertion.Len())

	serverAssertion.At(1, func(r http.Request) {
		assert.Equal(t, 2, len(r.Cookies()))

		assert.Equal(t, r.Cookies()[0].Name, cookies[0].Name)
		assert.Equal(t, r.Cookies()[0].Value, cookies[0].Value)
		assert.Equal(t, r.Cookies()[0].Expires, cookies[0].Expires)

		assert.Equal(t, r.Cookies()[1].Name, serverCookieName)
		assert.Equal(t, r.Cookies()[1].Value, serverCookieValue)
	})
}

func TestShouldForwardCookiesWhenJarIsDisabled(t *testing.T) {
	testServer, serverAssertion := test.NewServerWithAssertion(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	)
	defer testServer.Close()

	u, err := url.Parse(testServer.URL)
	assert.NoError(t, err)

	cookies := []*http.Cookie{
		{
			Name:  "a_cookie_name",
			Value: "a_cookie_value",
		},
	}

	c, err := client.NewClientFromConfig(
		100,
		nil,
		"",
		false,
		cookies,
		u,
	)
	assert.NoError(t, err)

	res, err := c.Get(testServer.URL)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assert.Equal(t, 1, serverAssertion.Len())

	serverAssertion.At(0, func(r http.Request) {
		assert.Equal(t, 1, len(r.Cookies()))

		assert.Equal(t, r.Cookies()[0].Name, cookies[0].Name)
		assert.Equal(t, r.Cookies()[0].Value, cookies[0].Value)
		assert.Equal(t, r.Cookies()[0].Expires, cookies[0].Expires)
	})
}

func TestShouldFailToCreateAClientWithInvalidSocks5Url(t *testing.T) {
	u := url.URL{Scheme: "potatoscheme"}

	c, err := client.NewClientFromConfig(
		100,
		&u,
		"",
		false,
		nil,
		nil,
	)
	assert.Nil(t, c)
	assert.Error(t, err)

	assert.Contains(t, err.Error(), "unknown scheme")
}