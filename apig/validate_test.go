package apig

import "testing"

func TestValidateForeignKey(t *testing.T) {
	model := &Model{
		Name: "User",
		Fields: []*Field{
			&Field{
				Name: "ID",
				Type: "uint",
			},
			&Field{
				Name: "Profile",
				Type: "*Profile",
			},
			&Field{
				Name: "ProfileID",
				Type: "uint",
			},
			&Field{
				Name: "CreatedAt",
				Type: "*time.Time",
			},
			&Field{
				Name: "UpdatedAt",
				Type: "*time.Time",
			},
		},
	}

	result := validateForeignKey(model.Fields, "Profile")
	if result != true {
		t.Fatalf("Incorrect result. expected: true, actual: %v", result)
	}

	result = validateForeignKey(model.Fields, "Nation")
	if result != false {
		t.Fatalf("Incorrect result. expected: false, actual: %v", result)
	}
}
