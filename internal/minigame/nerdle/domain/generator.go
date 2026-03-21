package domain

import (
	"fmt"
	"math/rand"
	"time"
)

func GenerateDailyFormula(date time.Time) string {
	seed := int64(date.Year())*10000 + int64(date.Month())*100 + int64(date.Day())
	rng := rand.New(rand.NewSource(seed))

	generators := []func(*rand.Rand) (string, bool){
		generateAddition,
		generateSubtraction,
		generateMultiplication,
		generateDivision,
		generateTwoOperators,
	}

	// Shuffle voor variatie
	rng.Shuffle(len(generators), func(i, j int) {
		generators[i], generators[j] = generators[j], generators[i]
	})

	for i := 0; i < 20; i++ {
		for _, gen := range generators {
			if formula, ok := gen(rng); ok && fitsInGrid(formula) {
				return formula
			}
		}
	}

	return "3+9-2=10" // fallback hihihi
}

func fitsInGrid(formula string) bool {
	return len(formula) <= 8
}

func generateAddition(rng *rand.Rand) (string, bool) {
	a := rng.Intn(99) + 1
	b := rng.Intn(99) + 1
	result := a + b

	formula := fmt.Sprintf("%d+%d=%d", a, b, result)

	if !fitsInGrid(formula) {
		return "", false
	}

	if err := ValidateFormula(formula); err != nil {
		return "", false
	}
	return formula, true
}

func generateSubtraction(rng *rand.Rand) (string, bool) {
	b := rng.Intn(99) + 1
	result := rng.Intn(99) + 1
	a := b + result

	formula := fmt.Sprintf("%d-%d=%d", a, b, result)

	if !fitsInGrid(formula) {
		return "", false
	}

	if err := ValidateFormula(formula); err != nil {
		return "", false
	}
	return formula, true
}

func generateMultiplication(rng *rand.Rand) (string, bool) {
	a := rng.Intn(20) + 1
	b := rng.Intn(12) + 1
	result := a * b

	formula := fmt.Sprintf("%d*%d=%d", a, b, result)

	if !fitsInGrid(formula) {
		return "", false
	}

	if err := ValidateFormula(formula); err != nil {
		return "", false
	}
	return formula, true
}

func generateDivision(rng *rand.Rand) (string, bool) {
	b := rng.Intn(12) + 1
	result := rng.Intn(20) + 1
	a := b * result

	formula := fmt.Sprintf("%d/%d=%d", a, b, result)

	if !fitsInGrid(formula) {
		return "", false
	}

	if err := ValidateFormula(formula); err != nil {
		return "", false
	}
	return formula, true
}

func generateTwoOperators(rng *rand.Rand) (string, bool) {
	ops := []string{"+", "-", "*", "/"}
	op1 := ops[rng.Intn(len(ops))]
	op2 := ops[rng.Intn(len(ops))]

	a := rng.Intn(9) + 1
	b := rng.Intn(9) + 1
	c := rng.Intn(9) + 1

	expr := fmt.Sprintf("%d%s%d%s%d", a, op1, b, op2, c)

	result, err := evalExpression(expr)
	if err != nil || result <= 0 || result != float64(int(result)) {
		return "", false
	}

	formula := fmt.Sprintf("%s=%d", expr, int(result))

	if !fitsInGrid(formula) {
		return "", false
	}

	if err := ValidateFormula(formula); err != nil {
		return "", false
	}

	return formula, true
}
