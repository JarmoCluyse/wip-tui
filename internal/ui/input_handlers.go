package ui

// removeLastCharacter removes the last character from the input field.
func (m Model) removeLastCharacter() Model {
	if len(m.InputField) > 0 {
		m.InputField = m.InputField[:len(m.InputField)-1]
	}
	return m
}

// appendCharacter appends a single character to the input field.
func (m Model) appendCharacter(char string) Model {
	if len(char) == 1 {
		m.InputField += char
	}
	return m
}
