package main

import (
	"encoding/json"
	"html/template"
	"log"
	"math"
	"net/http"
	"strconv"
)

// tmpl is our HTML template with embedded CSS and JavaScript.
var tmpl = template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html lang="uk">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Калькулятор</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            background-color: #e0e0e0;
            text-align: center;
            padding: 20px;
        }
        .container {
            background: white;
            padding: 20px;
            border-radius: 10px;
            box-shadow: 0px 0px 10px rgba(0, 0, 0, 0.1);
            display: inline-block;
        }
        input {
            width: 100px;
            padding: 5px;
            margin: 5px;
            text-align: center;
        }
        .result {
            background: #f8f8f8;
            padding: 15px;
            border-radius: 8px;
            box-shadow: 0px 0px 5px rgba(0, 0, 0, 0.1);
            margin-top: 20px;
        }
        button {
            font-size: 18px;
            padding: 10px 20px;
            margin: 10px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h2>Калькулятор</h2>
        <div>
            <label>naming: <input type="number" id="naming" value="0" step="1"></label>
            <label>coefN: <input type="number" id="coefN" value="0" step="0.01"></label>
            <label>coefP: <input type="number" id="coefP" value="0" step="0.01"></label>
            <label>strength: <input type="number" id="strength" value="0" step="0.01"></label>
            <label>quantity: <input type="number" id="quantity" value="0" step="0.01"></label>
            <label>nomP: <input type="number" id="nomP" value="0" step="0.01"></label>
            <label>coefU: <input type="number" id="coefU" value="0" step="0.01"></label>
            <label>coefRP: <input type="number" id="coefRP" value="0" step="0.01"></label>
        </div>
        <div>
            <button onclick="calculate()">Розрахувати</button>
        </div>
        <div class="result">
            <h3>Результати розрахунків</h3>
            <p>Розрахунковий струм: <span id="current">0.00</span></p>
            <p>Груповий коефіцієнт використання: <span id="groupUtilization">0.00</span></p>
            <p>Ефективна кількість ЕП: <span id="effectiveEP">0.00</span></p>
            <p>Активна потужність: <span id="activePower">0.00</span></p>
            <p>Реактивна потужність: <span id="reactivePower">0.00</span></p>
            <p>Повна потужність: <span id="fullPower">0.00</span></p>
        </div>
    </div>
    <script>
        async function calculate() {
            const data = {
                naming: parseFloat(document.getElementById('naming').value) || 0,
                coefN: parseFloat(document.getElementById('coefN').value) || 0,
                coefP: parseFloat(document.getElementById('coefP').value) || 0,
                strength: parseFloat(document.getElementById('strength').value) || 0,
                quantity: parseFloat(document.getElementById('quantity').value) || 0,
                nomP: parseFloat(document.getElementById('nomP').value) || 0,
                coefU: parseFloat(document.getElementById('coefU').value) || 0,
                coefRP: parseFloat(document.getElementById('coefRP').value) || 0
            };

            try {
                const response = await fetch('/calculate', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
                if (!response.ok) {
                    throw new Error("HTTP error " + response.status);
                }
                const result = await response.json();
                document.getElementById('current').textContent = result.current.toFixed(2);
                document.getElementById('groupUtilization').textContent = result.groupUtilization.toFixed(2);
                document.getElementById('effectiveEP').textContent = result.effectiveEP.toFixed(2);
                document.getElementById('activePower').textContent = result.activePower.toFixed(2);
                document.getElementById('reactivePower').textContent = result.reactivePower.toFixed(2);
                document.getElementById('fullPower').textContent = result.fullPower.toFixed(2);
            } catch (error) {
                console.error("Calculation error:", error);
            }
        }
    </script>
</body>
</html>
`))

// calcInput holds the JSON request payload.
type calcInput struct {
	Naming   float64 `json:"naming"`
	CoefN    float64 `json:"coefN"`
	CoefP    float64 `json:"coefP"`
	Strength float64 `json:"strength"`
	Quantity float64 `json:"quantity"`
	NomP     float64 `json:"nomP"`
	CoefU    float64 `json:"coefU"`
	CoefRP   float64 `json:"coefRP"`
}

// calcOutput holds the JSON response payload.
type calcOutput struct {
	Current          float64 `json:"current"`
	GroupUtilization float64 `json:"groupUtilization"`
	EffectiveEP      float64 `json:"effectiveEP"`
	ActivePower      float64 `json:"activePower"`
	ReactivePower    float64 `json:"reactivePower"`
	FullPower        float64 `json:"fullPower"`
}

// Calculation functions mimic the Android app logic.

// calculateCurrent computes (n * Pn) / (sqrt(3)*strength*quantity*nomP)
// where naming is converted to an integer.
func calculateCurrent(naming, coefP, strength, quantity, nomP float64) float64 {
	n := int(naming)
	denom := math.Sqrt(3) * strength * quantity * nomP
	if denom == 0 {
		return 0
	}
	return float64(n)*coefP / denom
}

// calculateGroupUtilizationCoefficient computes coefN / coefU.
func calculateGroupUtilizationCoefficient(coefN, coefU float64) float64 {
	if coefU == 0 {
		return 0
	}
	return coefN / coefU
}

// calculateEffectiveNumberEP computes (coefN^2) / coefRP.
func calculateEffectiveNumberEP(coefN, coefRP float64) float64 {
	if coefRP == 0 {
		return 0
	}
	return (coefN * coefN) / coefRP
}

// calculateActivePower computes coefN * coefU * coefP.
func calculateActivePower(coefN, coefU, coefP float64) float64 {
	return coefN * coefU * coefP
}

// calculateReactivePower computes coefU * coefP * coefRP.
func calculateReactivePower(coefU, coefP, coefRP float64) float64 {
	return coefU * coefP * coefRP
}

// calculateFullPower computes sqrt(activePower^2 + reactivePower^2).
func calculateFullPower(activePower, reactivePower float64) float64 {
	return math.Sqrt(activePower*activePower + reactivePower*reactivePower)
}

// indexHandler serves the main HTML page.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Помилка шаблону", http.StatusInternalServerError)
		log.Println("Template execute error:", err)
	}
}

// calculateHandler decodes the JSON request, performs calculations, and writes a JSON response.
func calculateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не дозволено", http.StatusMethodNotAllowed)
		return
	}

	var input calcInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Невірний запит", http.StatusBadRequest)
		log.Println("JSON decode error:", err)
		return
	}

	// Perform the calculations.
	current := calculateCurrent(input.Naming, input.CoefP, input.Strength, input.Quantity, input.NomP)
	groupUtilization := calculateGroupUtilizationCoefficient(input.CoefN, input.CoefU)
	effectiveEP := calculateEffectiveNumberEP(input.CoefN, input.CoefRP)
	activePower := calculateActivePower(input.CoefN, input.CoefU, input.CoefP)
	reactivePower := calculateReactivePower(input.CoefU, input.CoefP, input.CoefRP)
	fullPower := calculateFullPower(activePower, reactivePower)

	output := calcOutput{
		Current:          current,
		GroupUtilization: groupUtilization,
		EffectiveEP:      effectiveEP,
		ActivePower:      activePower,
		ReactivePower:    reactivePower,
		FullPower:        fullPower,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(output); err != nil {
		http.Error(w, "Помилка кодування JSON", http.StatusInternalServerError)
		log.Println("JSON encode error:", err)
	}
}

// main starts the HTTP server and registers handlers.
func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/calculate", calculateHandler)
	port := 9090
	log.Println("Server starting on port " + strconv.Itoa(port))
	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		log.Fatal("ListenAndServe error:", err)
	}
}