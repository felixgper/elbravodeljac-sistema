package postgres

import "context"

func (s *Store) Brands(ctx context.Context) ([]string, error) {
	rows, err := s.Pool.Query(ctx, `SELECT nombre FROM public.marcas_vehiculos ORDER BY public.jac_normalize(nombre),nombre`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}
