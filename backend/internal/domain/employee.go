package domain

type EmployeeInput struct {
	Name   *string `json:"name"`
	Type   *string `json:"employeeType"`
	Active *bool   `json:"active"`
}

func (v EmployeeInput) Validate(create bool) error {
	if create && (v.Name == nil || v.Type == nil) {
		return Invalid("employee", "Nombre y tipo de empleado son obligatorios.")
	}
	for _, f := range []struct {
		key   string
		value *string
		max   int
	}{{"name", v.Name, 150}, {"employeeType", v.Type, 40}} {
		if f.value != nil {
			if err := text(f.key, *f.value, f.max, true); err != nil {
				return err
			}
		}
	}
	return nil
}
