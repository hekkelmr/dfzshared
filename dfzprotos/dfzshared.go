/*
Copyright Rob Hekkelman 2017 All Rights Reserved.
*/

package dfzprotos

import (
	"time"
)

// EIVoorloopRecord ... The header record of a healthclaim
type EIVoorloopRecord struct {
	AGBServicebureau        string `json:"AGBServicebureau"`
	AGBZorgverlener         string `json:"AGBZorgverlener"`
	AGBPraktijk             string `json:"AGBPraktijk"`
	AGBInstelling           string `json:"AGBInstelling"`
	IdentificatieBetalenAan string `json:"IdentificatieBetalenAan"`
	ReferentieBehandeling   string `json:"ReferentieBehandeling"`
}

// EIVerzekerdeRecord ... De patient
type EIVerzekerdeRecord struct {
	Bsncode string `json:"Bsncode"`
	Naam    string `json:"Naam"`
}

// EIPrestatieRecord ... Een detail van een behandeling
type EIPrestatieRecord struct {
	Prestatiecodelijst string    `json:"Prestatiecodelijst"`
	Prestatiecode      string    `json:"Prestatiecode"`
	TariefPrestatie    int64     `json:"TariefPrestatie"`
	BerekendBedrag     int64     `json:"BerekendBedrag"`
	DatumPrestatie     time.Time `json:"DatumPrestatie"`
}

type PrestatieResultaat struct {
	PrestatieRecord EIPrestatieRecord `json:"PrestatieRecord"`
	Omschrijving    string            `json:"Omschrijving"`
	Bericht         string            `json:"Bericht"`
}

// Declaratie ... Combined healtclaim structure
type Declaratie struct {
	Bsncode          string              `json:"Bsncode"`
	Area             string              `json:"Area"`
	Year             string              `json:"Year"`
	Voorlooprecord   EIVoorloopRecord    `json:"Voorlooprecord"`
	Verzekerderecord EIVerzekerdeRecord  `json:"Verzekerderecord"`
	Prestatierecords []EIPrestatieRecord `json:"Prestatierecords"`
	GeoLocatie       GeoLocatie          `json:"GeoLocatie"`
}

type GeoLocatie struct {
	Lat string `json:"Lat"`
	Lon string `json:"Lon"`
}

// ContractStatus ...
// De status van een polis
type ContractStatus struct {
	Tegoed     int64      `json:"Tegoed"`
	Eenheid    string     `json:"Eenheid"`
	Declaratie Declaratie `json:"Declaratie"`
}

// PolisJaar ...
// De naam van een polis voor een bepaald jaar
type PolisJaar struct {
	Jaar              string `json:"Jaar"`
	PolisContractnaam string `json:"PolisContractnaam"`
}

// AfgeslotenPolisstatus ...
// De status van een sfgesloten polis
type AfgeslotenPolisstatus struct {
	AfgeslotenPolis AfgeslotenPolis `json:"AfgeslotenPolis"`
	ContractStatus  ContractStatus  `json:"ContractStatus"`
}

// AfgeslotenPolis ...
// De polis van een persoon in een bepaald jaar
type AfgeslotenPolis struct {
	Bsncode      string    `json:"Bsncode"`
	Verstrekking string    `json:"Verstrekking"`
	UzoviCode    string    `json:"Uzovicode"`
	PolisJaar    PolisJaar `json:"PolisJaar"`
}

// Polisafspraak ...
type Polisafspraak struct {
	UzoviCode             string
	ContractCode          string
	Verstrekking          string
	MaxAantalPerJaar      int64
	Eenheid               string
	Factor                float32
	VerzekeraarRepository string
	GebruikLocatieCheck   bool
}

// Persoon ...
type Persoon struct {
	Bsncode          string    `json:"Bsncode"`
	Naam             string    `json:"Naam"`
	Geboortedatum    time.Time `json:"Geboortedatum"`
	Overlijdensdatum time.Time `json:"Overlijdensdatum"`
	WalletID         string    `json:"WalletID"`
	ID               string    `json:"ID"`
}

type Zorgaanbieder struct {
	Agbcode     string       `json:"Agbcode"`
	Naam        string       `json:"Naam"`
	Type        string       `json:"Type"`
	Soort       string       `json:"Soort"`
	WalletID    string       `json:"WalletID"`
	ID          string       `json:"ID"`
	GeoLocaties []GeoLocatie `json:"GeoLocaties"`
}

type Verzekeraar struct {
	Uzovicode string `json:"Uzovicode"`
	Naam      string `json:"Naam"`
	WalletID  string `json:"WalletID"`
	ID        string `json:"ID"`
}

type Behandeling struct {
	BehandelingID string        `json:"BehandelingID"`
	Patient       Persoon       `json:"Patient"`
	Indiener      Zorgaanbieder `json:"Indiener"`
	Diagnose      string        `json:"Diagnose"`
	Behandeling   string        `json:"Behandeling"`
	Claim         Declaratie    `json:"Claim"`
	VorigeID      string        `json:"VorigeID"`
	Referentie    string        `json:"Referentie"`
	Uitgevoerd    string        `json:"Uitgevoerd"`
}

// type AssignedPolicies struct {
// 	years []PolicyYear `json:"years"`
// }

// Retourbericht ... De uitkomst van een ingediende claim
type Retourbericht struct {
	AgbCode          string               `json:"AgbCode"`
	Retourcode       string               `json:"Retourcode"`
	Restant          int64                `json:"Restant"`
	Vergoed          int64                `json:"Vergoed"`
	RestantEenheid   string               `json:"RestantEenheid"`
	Bijbetalen       int64                `json:"Bijbetalen"`
	Bericht          string               `json:"Bericht"`
	Prestatierecords []PrestatieResultaat `json:"Prestatierecords"`
}

type ZorgaanbiedersContractRouter struct {
	UzoviCode string `json:"UzoviCode"`
	RESTUrl   string `json:"RESTUrl"`
}

type WalletTransactie struct {
	Van    string `json:"Van"`
	Bedrag int64  `json:"Bedrag"`
	Data   string `json:"Data"`
}

type CurecoinWallet struct {
	ID                string           `json:"ID"`
	Saldo             int64            `json:"Saldo"`
	LaatsteTransactie WalletTransactie `json:"LaatsteTransactie"`
}

type BehandelingToegang struct {
	Verlener       string    `json:"Verlener"`
	Geautoriseerde string    `json:"Geautoriseerde"`
	Vervalt        time.Time `json:"Vervalt"`
}

type Agb struct {
	Agbcode string `json:"agbcode"`
	Naam    string `json:"Naam"`
	Type    string `json:"Type"`
	Soort   string `json:"Soort"`
}

type Contractprestatie struct {
	Prestatiecodelijst string `json:"Prestatiecodelijst"`
	Prestatiecode      string `json:"Prestatiecode"`
	TariefPrestatie    int64  `json:"TariefPrestatie"`
	Omschrijving       string `json:"Omschrijving"`
	Percentage         int64  `json:"Percentage"`
	Herkomst           string `json:"Herkomst"`
}
