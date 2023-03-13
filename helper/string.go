package helper

import (
	"fmt"
	"strings"

	"github.com/LalatinaHub/LatinaApi/common/account/converter"
	"github.com/LalatinaHub/LatinaSub-go/db"
)

func MakeVPNMessage(account db.DBScheme) string {
	message := []string{}
	message = append(message, fmt.Sprintf("<code>PROTOCOL      : %s</code>", account.VPN))
	message = append(message, fmt.Sprintf("<code>REMARKS       : %s</code>", account.Remark))
	message = append(message, fmt.Sprintf("<code>CC            : %s</code>", account.CountryCode))
	message = append(message, fmt.Sprintf("<code>REGION        : %s</code>", account.Region))
	message = append(message, fmt.Sprintf("<code>SERVER        : %s</code>", account.Server))
	message = append(message, fmt.Sprintf("<code>HOST          : %s</code>", account.Host))
	message = append(message, fmt.Sprintf("<code>SNI           : %s</code>", account.SNI))
	message = append(message, fmt.Sprintf("<code>PORT          : %d</code>", account.ServerPort))
	message = append(message, fmt.Sprintf("<code>UUID          : %s</code>", account.UUID))
	message = append(message, fmt.Sprintf("<code>PASSWORD      : %s</code>", account.Password))
	message = append(message, fmt.Sprintf("<code>ISP           : %s</code>", account.Org))
	message = append(message, fmt.Sprintf("<code>MODE          : %s</code>", account.ConnMode))
	message = append(message, fmt.Sprintf("<code>TLS           : %t</code>", account.TLS))
	message = append(message, fmt.Sprintf("<code>Network       : %s</code>", account.Transport))
	message = append(message, fmt.Sprintf("<code>PATH          : %s</code>", account.Path))
	message = append(message, fmt.Sprintf("<code>SERVICE NAME  : %s</code>", account.ServiceName))
	message = append(message, fmt.Sprintf("<code>%s</code>", "-------------------------"))
	message = append(message, fmt.Sprintf("<code>%s</code>", converter.ToRaw([]db.DBScheme{account})))
	message = append(message, fmt.Sprintf("<code>%s</code>", "-------------------------"))

	return strings.Join(message, "\n")
}
