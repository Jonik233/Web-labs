package main

import (
	"html/template"
	"math"
	"net/http"
	"strconv"
)

var tmpl = template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html lang="uk">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Калькулятор струму КЗ</title>
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
            width: 80px;
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
        <h2>Калькулятор струму короткого замикання</h2>
        <label>U (Фазна напруга): <input type="number" id="U" value="0" step="0.01"></label>
        <label>Z (Повний опір): <input type="number" id="Z" value="0" step="0.01"></label>
        <label>R (Активний опір): <input type="number" id="R" value="0" step="0.01"></label>
        <label>X (Реактивний опір): <input type="number" id="X" value="0" step="0.01"></label>
        <label>t (Час): <input type="number" id="T" value="0" step="0.01"></label>
        <label>kf (Коефіцієнт форми): <input type="number" id="KF" value="0" step="0.01"></label>
        <label>kThermal (Термічна стійкість): <input type="number" id="KThermal" value="0" step="0.01"></label>
        <label>kDyn (Динамічна стійкість): <input type="number" id="KDyn" value="0" step="0.01"></label>

        <div>
            <button onclick="calculate()">Розрахувати</button>
        </div>

        <div class="result">
            <h3>Результати розрахунків</h3>
            <p>Трифазний струм: <span id="threePhaseCurrent">0.00</span></p>
            <p>Однофазний струм: <span id="singlePhaseCurrent">0.00</span></p>
            <p>Термічна стійкість: <span id="isThermalStable">-</span></p>
            <p>Динамічна стійкість: <span id="isDynamicStable">-</span></p>
        </div>
    </div>

    <script>
        async function calculate() {
            const data = {
                U: parseFloat(document.getElementById('U').value) || 0,
                Z: parseFloat(document.getElementById('Z').value) || 0,
                R: parseFloat(document.getElementById('R').value) || 0,
                X: parseFloat(document.getElementById('X').value) || 0,
                T: parseFloat(document.getElementById('T').value) || 0,
                KF: parseFloat(document.getElementById('KF').value) || 0,
                KThermal: parseFloat(document.getElementById('KThermal').value) || 0,
                KDyn: parseFloat(document.getElementById('KDyn').value) || 0
            };

            const response = await fetch('/calculate', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });

            const result = await response.json();
            document.getElementById('threePhaseCurrent').textContent = result.threePhaseCurrent.toFixed(2);
            document.getElementById('singlePhaseCurrent').textContent = result.singlePhaseCurrent.toFixed(2);
            document.getElementById('isThermalStable').textContent = result.isThermalStable ? "Так" : "Ні";
            document.getElementById('isDynamicStable').textContent = result.isDynamicStable ? "Так" : "Ні";
        }
    </script>

</body>
</html>
`))

// Functions for calculations
func calculateThreePhaseCurrent(U, Z float64) float64 {
	if Z == 0 {
		return 0
	}
	return U / Z
}

func calculateSinglePhaseCurrent(U, R, X float64) float64 {
	Z := math.Sqrt(R*R + X*X)
	if Z == 0 {
		return 0
	}
	return U / Z
}

func checkThermalStability(I, k, t float64) bool {
	return I*I*t <= k
}

func checkDynamicStability(Ipeak, kdyn float64) bool {
	return Ipeak <= kdyn
}

func calculatePeakCurrent(I, kf float64) float64 {
	return I * kf
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.Execute(w, nil)
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не дозволено", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	U, _ := strconv.ParseFloat(r.FormValue("U"), 64)
	Z, _ := strconv.ParseFloat(r.FormValue("Z"), 64)
	R, _ := strconv.ParseFloat(r.FormValue("R"), 64)
	X, _ := strconv.ParseFloat(r.FormValue("X"), 64)
	T, _ := strconv.ParseFloat(r.FormValue("T"), 64)
	KF, _ := strconv.ParseFloat(r.FormValue("KF"), 64)
	KThermal, _ := strconv.ParseFloat(r.FormValue("KThermal"), 64)
	KDyn, _ := strconv.ParseFloat(r.FormValue("KDyn"), 64)

	threePhaseCurrent := calculateThreePhaseCurrent(U, Z)
	singlePhaseCurrent := calculateSinglePhaseCurrent(U, R, X)
	isThermalStable := checkThermalStability(threePhaseCurrent, KThermal, T)
	peakCurrent := calculatePeakCurrent(threePhaseCurrent, KF)
	isDynamicStable := checkDynamicStability(peakCurrent, KDyn)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{
		"threePhaseCurrent": ` + strconv.FormatFloat(threePhaseCurrent, 'f', 2, 64) + `,
		"singlePhaseCurrent": ` + strconv.FormatFloat(singlePhaseCurrent, 'f', 2, 64) + `,
		"isThermalStable": ` + strconv.FormatBool(isThermalStable) + `,
		"isDynamicStable": ` + strconv.FormatBool(isDynamicStable) + `
	}`))
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/calculate", calculateHandler)
	http.ListenAndServe(":8080", nil)
}