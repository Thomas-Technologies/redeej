const int NUM_SLIDERS = 5;
const int analogInputs[NUM_SLIDERS] = {A3, A2, A1, A0, A10};
const int NUM_BUTTONS = 5;
const int digitalInputs[NUM_BUTTONS] = {5, 7, 6, 8, 9};

// const int NUM_SLIDERS = 2;
// const int analogInputs[NUM_SLIDERS] = { A3, A2};
// const int NUM_BUTTONS = 2;
// const int digitalInputs[NUM_BUTTONS] = {5, 6};

int analogSliderValues[NUM_SLIDERS];
int buttonValues[NUM_BUTTONS];

void setup() {
  for (int i = 0; i < NUM_SLIDERS; i++) {
    pinMode(analogInputs[i], INPUT);
  }

  for (int i = 0; i < NUM_BUTTONS; i++) {
    pinMode(digitalInputs[i], INPUT_PULLUP);
  }

  Serial.begin(9600);
}

void loop() {
  updateSliderValues();
  updateButtonValues();
  sendValues(); // Actually send data (all the time)
  // printSliderValues(); // For debug
  delay(10);
}

void updateSliderValues() {
  for (int i = 0; i < NUM_SLIDERS; i++) {
     analogSliderValues[i] = analogRead(analogInputs[i]);
  }
}

void updateButtonValues() {
  for (int i = 0; i < NUM_BUTTONS; i++) {
     buttonValues[i] = digitalRead(digitalInputs[i]);
  }
}

void sendValues() {
  String builtString = String("");

  for (int i = 0; i < NUM_SLIDERS; i++) {
    builtString += String((int)analogSliderValues[i]);

    if (i < NUM_SLIDERS - 1) {
      builtString += String("|");
    }
  }

  if (NUM_BUTTONS > 0) {
    builtString += String("||");
    for (int i = 0;i < NUM_BUTTONS; i++) {
      builtString += String((int) buttonValues[i]);

      if (i < NUM_BUTTONS - 1) {
        builtString += String("|");
      }
    }
  }

  Serial.println(builtString);
}

void printSliderValues() {
  for (int i = 0; i < NUM_SLIDERS; i++) {
    String printedString = String("Slider #") + String(i + 1) + String(": ") + String(analogSliderValues[i]) + String(" mV");
    Serial.write(printedString.c_str());

    if (i < NUM_SLIDERS - 1) {
      Serial.write(" | ");
    } else {
      Serial.write("\n");
    }
  }
}
