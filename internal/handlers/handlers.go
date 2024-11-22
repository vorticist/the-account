package handlers

import (
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/lithammer/shortuuid/v4"
	"html/template"
	"regexp"
	"strings"
	"vortex.studio/account/internal/repo"
	"vortex.studio/account/internal/structs"
	"vortex.studio/account/internal/utils"
)

var (
	store = sessions.NewCookieStore([]byte("super-secret-key"))
)

var templateFuncs = template.FuncMap{
	"makeURLSafe": makeURLSafe,
	"getItemVals": getItemVals,
}

type Handler struct {
	venueRepo        *repo.VenueRepository
	activeTablesRepo *repo.ActiveTablesRepository
}

func makeURLSafe(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace spaces with hyphens
	name = strings.ReplaceAll(name, " ", "-")

	// Remove special characters and keep only alphanumeric and hyphens
	reg := regexp.MustCompile("[^a-z0-9-]+")
	name = reg.ReplaceAllString(name, "")

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile("-+")
	name = reg.ReplaceAllString(name, "-")

	// Trim hyphens from start and end
	name = strings.Trim(name, "-")

	return name
}

func generateTableCodes(venue *structs.Venue, howMany int) error {
	venue.TableCodes = []structs.TableCode{}
	for i := 0; i < howMany; i++ {
		u := shortuuid.New()
		tableCodeUrl := fmt.Sprintf("http://localhost:9090/table/%s", u)
		qrCode, err := utils.GenerateQRCodeBase64(tableCodeUrl)
		if err != nil {
			return err
		}
		venue.TableCodes = append(venue.TableCodes, structs.TableCode{
			Code:    u,
			CodeUrl: tableCodeUrl,
			Base64:  qrCode,
		})
	}
	return nil
}

func getItemVals(item structs.MenuItem) string {
	return fmt.Sprintf(`{"name": "%s", "description": "%s", "price": %v, "amount": 1}`, item.Name, item.Description, item.Price)
}
