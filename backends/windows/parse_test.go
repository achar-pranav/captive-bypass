package windows

import "testing"

const enUSConnected = `There is 1 interface on the system:

    Name                   : WiFi
    Description            : Intel(R) Wi-Fi 6 AX201 160MHz
    GUID                   : 8a3b1c2d-1111-2222-3333-abcdefabcdef
    Physical address       : d4:6a:6a:11:22:33
    State                  : connected
    SSID                   : ELEMENT BLOCK
    BSSID                  : ce:82:a9:d5:35:8b
    Network type           : Infrastructure
    Radio type             : 802.11ax
    Authentication         : WPA2-Personal
    Cipher                 : CCMP
    Connection mode        : Auto Connect
    Channel                : 44
    Receive rate (Mbps)    : 1201
    Transmit rate (Mbps)   : 1201
    Signal                 : 86%
    Profile                : ELEMENT BLOCK

    Hosted network status  : Not started
`

func TestParseEnglishConnected(t *testing.T) {
	s := parseInterfaces(enUSConnected)
	if s.SSID != "ELEMENT BLOCK" || s.BSSID != "ce:82:a9:d5:35:8b" || s.Signal != 86 || !s.Up {
		t.Errorf("snapshot = %+v", s)
	}
}

func TestParseSSIDWithColons(t *testing.T) {
	out := enUSConnected[:len(enUSConnected)]
	s := parseInterfaces(out)
	_ = s
	fixture := `
    Name                   : WiFi
    Physical address       : d4:6a:6a:11:22:33
    State                  : connected
    SSID                   : PESU:STAFF:5G
    BSSID                  : ce:82:a9:d5:35:8b
    Signal                 : 72%
`
	got := parseInterfaces(fixture)
	if got.SSID != "PESU:STAFF:5G" {
		t.Errorf("SSID = %q, want %q", got.SSID, "PESU:STAFF:5G")
	}
}

const deDEConnected = `Es besteht eine Verbindung mit 1 Schnittstelle:

    Name                   : WLAN
    Description            : Intel(R) Wi-Fi 6 AX201 160MHz
    GUID                   : 8a3b1c2d-1111-2222-3333-abcdefabcdef
    Physische Adresse      : d4:6a:6a:11:22:33
    Status                 : verbunden
    SSID                   : Campus-Netz
    BSSID                  : aa:bb:cc:00:11:22
    Netzwerktyp            : Infrastruktur
    Funktyp                : 802.11ax
    Authentifizierung      : WPA2-Personal
    Verschlüsselung        : CCMP
    Verbindungsmodus       : Automatisch verbinden
    Kanal                  : 6
    Empfangsrate (MBit/s)  : 1201
    Übertragungsrate (MBit/s): 1201
    Signal                 : 64%
    Profil                 : Campus-Netz
`

func TestParseGermanConnected(t *testing.T) {
	s := parseInterfaces(deDEConnected)
	if s.SSID != "Campus-Netz" || s.BSSID != "aa:bb:cc:00:11:22" || s.Signal != 64 || !s.Up {
		t.Errorf("snapshot = %+v", s)
	}
}

const frFRConnected = `Il y a 1 interface sur le système :

    Nom                     : WiFi
    Description             : Intel(R) Wi-Fi 6 AX201 160MHz
    GUID                    : {guid}
    Adresse physique        : d4:6a:6a:11:22:33
    État                    : connecté
    SSID                    : Reseau:Campus
    BSSID                   : ab:cd:ef:01:02:03
    Type de réseau          : Infrastructure
    Type de radio           : 802.11ax
    Authentification        : WPA2-Personnel
    Chiffrement             : CCMP
    Mode de connexion       : Connexion automatique
    Canal                   : 36
    Taux de réception (Mb/s): 1201
    Signal                  : 91%
    Profil                  : Reseau:Campus
`

func TestParseFrenchConnected(t *testing.T) {
	s := parseInterfaces(frFRConnected)
	if s.SSID != "Reseau:Campus" || s.BSSID != "ab:cd:ef:01:02:03" || s.Signal != 91 || !s.Up {
		t.Errorf("snapshot = %+v", s)
	}
}

const esESConnected = `Hay 1 interfaz en el sistema:

    Nombre                    : Wi-Fi
    Descripción               : Intel(R) Wi-Fi 6 AX201 160MHz
    GUID                      : {guid}
    Dirección física          : d4:6a:6a:11:22:33
    Estado                    : conectado
    SSID                      : Red:Facultad
    BSSID                     : 0f:1e:2d:3c:4b:5a
    Tipo de red               : Infraestructura
    Tipo de radio             : 802.11ac
    Autenticación             : WPA2-Personal
    Cifrado                   : CCMP
    Modo de conexión          : Conexión automática
    Canal                     : 11
    Velocidad de recepción    : 866
    Señal                     : 38%
    Perfil                    : Red:Facultad
`

func TestParseSpanishConnected(t *testing.T) {
	s := parseInterfaces(esESConnected)
	if s.SSID != "Red:Facultad" || s.BSSID != "0f:1e:2d:3c:4b:5a" || s.Signal != 38 || !s.Up {
		t.Errorf("snapshot = %+v", s)
	}
}

const enUSDisconnected = `There is 1 interface on the system:

    Name                   : WiFi
    Description            : Intel(R) Wi-Fi 6 AX201 160MHz
    GUID                   : 8a3b1c2d-1111-2222-3333-abcdefabcdef
    Physical address       : d4:6a:6a:11:22:33
    State                  : disconnected
    Hosted network status  : Not started
`

func TestParseDisconnected(t *testing.T) {
	s := parseInterfaces(enUSDisconnected)
	if s.Up {
		t.Errorf("snapshot = %+v, want Up=false", s)
	}
}

func TestParseHiddenSSIDStillUp(t *testing.T) {
	out := `
    Name                   : WiFi
    Physical address       : d4:6a:6a:11:22:33
    State                  : connected
    BSSID                  : ce:82:a9:d5:35:8b
    Signal                 : 50%
`
	s := parseInterfaces(out)
	if !s.Up || s.BSSID != "ce:82:a9:d5:35:8b" {
		t.Errorf("snapshot = %+v, want Up with hidden SSID", s)
	}
}

func TestParsePercentGarbage(t *testing.T) {
	if got := parsePercent("not a percent"); got != -1 {
		t.Errorf("parsePercent = %d, want -1", got)
	}
}
