package domain

import (
	_ "embed"
	"encoding/json"
	"sort"
	"strings"
)

//go:embed permission-policy.json
var policyJSON []byte

type Matrix map[string][]string
type Policy struct {
	Modules  map[string][]string `json:"modules"`
	Defaults map[string]Matrix   `json:"defaults"`
}

func PermissionPolicy() Policy {
	var p Policy
	if err := json.Unmarshal(policyJSON, &p); err != nil {
		panic("invalid embedded permission policy")
	}
	return p
}
func (m Matrix) Codes() ([]string, error) {
	allowed := PermissionPolicy().Modules
	codes := []string{}
	for module, actions := range m {
		valid, ok := allowed[module]
		if !ok {
			return nil, Invalid("permissions", "Módulo desconocido.")
		}
		seen := map[string]bool{}
		for _, action := range actions {
			found := false
			for _, a := range valid {
				if action == a {
					found = true
				}
			}
			if !found || seen[action] {
				return nil, Invalid("permissions", "Acción desconocida o duplicada.")
			}
			seen[action] = true
			codes = append(codes, module+"."+action)
		}
		if len(actions) > 0 && !seen["ver"] {
			return nil, Invalid("permissions", "Las acciones requieren permiso ver en su módulo.")
		}
	}
	sort.Strings(codes)
	return codes, nil
}

type MatrixInput struct {
	Permissions Matrix `json:"permissions"`
}
type UserInput struct {
	EmployeeID   *string `json:"employeeId"`
	Name         *string `json:"name"`
	Username     *string `json:"username"`
	Email        *string `json:"email"`
	Password     *string `json:"password"`
	Active       *bool   `json:"active"`
	RoleID       *string `json:"roleId"`
	RemoveRoleID *string `json:"removeRoleId"`
}

func (v UserInput) Validate(create bool) error {
	if create && (v.Name == nil || v.Username == nil || v.Password == nil || v.RoleID == nil) {
		return Invalid("request", "Nombre, usuario, contraseña y roleId son obligatorios.")
	}
	for _, f := range []struct {
		key      string
		value    *string
		max      int
		required bool
	}{{"name", v.Name, 150, true}, {"username", v.Username, 60, true}, {"email", v.Email, 150, false}} {
		if f.value != nil {
			if err := text(f.key, *f.value, f.max, f.required); err != nil {
				return err
			}
		}
	}
	if v.Username != nil && len(*v.Username) < 3 {
		return Invalid("username", "Mínimo tres caracteres.")
	}
	if v.Email != nil && *v.Email != "" && (!strings.Contains(*v.Email, "@") || strings.ContainsAny(*v.Email, " \r\n")) {
		return Invalid("email", "Correo inválido.")
	}
	if v.Password != nil && (len(*v.Password) < 12 || len(*v.Password) > 72) {
		return Invalid("password", "Usa entre 12 y 72 bytes para bcrypt.")
	}
	for _, f := range []struct {
		key   string
		value *string
	}{{"roleId", v.RoleID}, {"removeRoleId", v.RemoveRoleID}, {"employeeId", v.EmployeeID}} {
		if f.value != nil {
			if _, err := ID(*f.value); err != nil {
				return Invalid(f.key, "Rol inválido.")
			}
		}
	}
	if create && v.RemoveRoleID != nil {
		return Invalid("removeRoleId", "No corresponde retirar roles al crear.")
	}
	if v.RoleID != nil && v.RemoveRoleID != nil && *v.RoleID == *v.RemoveRoleID {
		return Invalid("roleId", "No agregues y retires el mismo rol.")
	}
	return nil
}

type ProviderInput struct {
	Name     *string `json:"name"`
	Document *string `json:"document"`
	RUC      *string `json:"ruc"`
	Phone    *string `json:"phone"`
	Contact  *string `json:"contact"`
	Parts    *string `json:"parts"`
}

func (v ProviderInput) Validate() error {
	if v.Parts != nil && *v.Parts != "" {
		return Invalid("parts", "Especialidad no tiene columna equivalente. Usa referencias explícitas de repuestos/costo o acuerda una futura migración; no se descartará este texto.")
	}
	if v.RUC != nil && v.Document != nil && *v.RUC != *v.Document {
		return Invalid("document", "document y ruc deben coincidir si se envían ambos.")
	}
	for _, f := range []struct {
		key      string
		value    *string
		max      int
		required bool
	}{{"name", v.Name, 150, true}, {"document", v.Document, 20, true}, {"ruc", v.RUC, 20, true}, {"phone", v.Phone, 20, false}, {"contact", v.Contact, 120, false}} {
		if f.value != nil {
			if err := text(f.key, *f.value, f.max, f.required); err != nil {
				return err
			}
		}
	}
	return nil
}

type CatalogInput struct {
	Name           string `json:"name"`
	ReferencePrice string `json:"referencePrice"`
	Stock          string `json:"stock"`
	Code           string `json:"code"`
	Description    string `json:"description"`
	Active         *bool  `json:"active"`
}

func (v *CatalogInput) Validate(resource string) error {
	if resource == "unidades-auxilio" {
		if v.Name != "" || v.ReferencePrice != "" || v.Stock != "" {
			return Invalid("request", "Una unidad utiliza code/description, no precios o stock.")
		}
		if err := text("code", v.Code, 30, true); err != nil {
			return err
		}
		return text("description", v.Description, 150, false)
	}
	max := 100
	if resource == "repuestos" {
		max = 120
	}
	if err := text("name", v.Name, max, true); err != nil {
		return err
	}
	n, err := Money(v.ReferencePrice, "referencePrice")
	if err != nil {
		return err
	}
	v.ReferencePrice = FormatMileage(n)
	if v.Code != "" || v.Description != "" {
		return Invalid("request", "code/description solo corresponden a unidades de auxilio.")
	}
	if resource == "repuestos" {
		if v.Stock == "" {
			v.Stock = "0.00"
		}
		n, err = Money(v.Stock, "stock")
		if err != nil {
			return err
		}
		v.Stock = FormatMileage(n)
	} else if v.Stock != "" {
		return Invalid("stock", "Un servicio no tiene stock.")
	}
	return nil
}

type AuxilioInput struct {
	ClientID     string `json:"clientId"`
	VehicleID    string `json:"vehicleId"`
	TechnicianID string `json:"technicianId"`
	UnitID       string `json:"unitId"`
	ReceptionID  string `json:"receptionId"`
	Location     string `json:"location"`
	Incident     string `json:"incident"`
	Description  string `json:"description"`
	State        string `json:"state"`
}

func (v AuxilioInput) Validate(create bool) error {
	for _, f := range []struct {
		field, id string
		required  bool
	}{{"clientId", v.ClientID, create}, {"vehicleId", v.VehicleID, create}, {"technicianId", v.TechnicianID, create}, {"unitId", v.UnitID, create}, {"receptionId", v.ReceptionID, false}} {
		if f.required || f.id != "" {
			if _, err := ID(f.id); err != nil {
				return Invalid(f.field, "ID obligatorio/inválido.")
			}
		}
	}
	if create {
		for _, e := range []error{text("location", v.Location, 300, true), text("incident", v.Incident, 60, true), text("description", v.Description, 4000, true)} {
			if e != nil {
				return e
			}
		}
	} else if v.ClientID != "" || v.VehicleID != "" || v.Location != "" || v.Incident != "" || v.Description != "" || v.ReceptionID != "" {
		return Invalid("request", "PATCH de auxilio solo admite técnico, unidad y estado.")
	}
	if v.State != "" && v.State != "pendiente" && v.State != "en_atencion" && v.State != "finalizado" && v.State != "cancelado" {
		return Invalid("state", "Estado inválido.")
	}
	return nil
}

type PartLinkInput struct {
	ReferenceCost string `json:"referenceCost"`
}
