import { describe, it, expect } from 'vitest'
import {
  detectDelimiter,
  splitCsvLine,
  parseUsersCsv
} from '../utils/csvParser'

describe('CSV Parser Utility', () => {
  describe('detectDelimiter', () => {
    it('detects comma delimiter', () => {
      const csv = 'email,firstName,lastName\nuser@encertia.cat,Marc,Rovira'
      expect(detectDelimiter(csv)).toBe(',')
    })

    it('detects semicolon delimiter', () => {
      const csv = 'email;nom;cognoms\nuser@encertia.cat;Marc;Rovira'
      expect(detectDelimiter(csv)).toBe(';')
    })

    it('detects tab delimiter', () => {
      const csv = 'email\tnom\tcognoms\nuser@encertia.cat\tMarc\tRovira'
      expect(detectDelimiter(csv)).toBe('\t')
    })
  })

  describe('splitCsvLine', () => {
    it('splits standard comma-separated cells', () => {
      const line = 'anna@encertia.cat, Anna , Serra '
      expect(splitCsvLine(line, ',')).toEqual(['anna@encertia.cat', 'Anna', 'Serra'])
    })

    it('handles quoted strings with commas inside', () => {
      const line = 'user@encertia.cat,"Serra, Costa",Anna'
      expect(splitCsvLine(line, ',')).toEqual(['user@encertia.cat', 'Serra, Costa', 'Anna'])
    })

    it('handles escaped quotes inside quoted strings', () => {
      const line = 'user@encertia.cat,"Marc ""El Gran""",Rovira'
      expect(splitCsvLine(line, ',')).toEqual(['user@encertia.cat', 'Marc "El Gran"', 'Rovira'])
    })
  })

  describe('parseUsersCsv', () => {
    it('returns error on empty CSV text', () => {
      const result = parseUsersCsv('')
      expect(result.errors.length).toBeGreaterThan(0)
      expect(result.totalCount).toBe(0)
    })

    it('parses valid CSV with standard English headers', () => {
      const csv = `email,firstName,lastName,password
laia.sole@encertia.cat,Laia,Sole,SecretPass123!
pol.vila@encertia.cat,Pol,Vila,Provis2026!`

      const result = parseUsersCsv(csv, 'student')

      expect(result.totalCount).toBe(2)
      expect(result.validCount).toBe(2)
      expect(result.invalidCount).toBe(0)
      expect(result.validItems).toHaveLength(2)
      expect(result.validItems[0]).toEqual({
        email: 'laia.sole@encertia.cat',
        firstName: 'Laia',
        lastName: 'Sole',
        role: 'student',
        password: 'SecretPass123!'
      })
      expect(result.validItems[1].email).toBe('pol.vila@encertia.cat')
    })

    it('parses Catalan headers correctly (nom, cognoms, correu)', () => {
      const csv = `correu;nom;cognoms;rol
maria.perez@encertia.cat;Maria;Perez;professor
jordi.cases@encertia.cat;Jordi;Cases;alumne`

      const result = parseUsersCsv(csv)

      expect(result.validCount).toBe(2)
      expect(result.validItems[0].role).toBe('teacher')
      expect(result.validItems[1].role).toBe('student')
    })

    it('flags invalid rows with appropriate error messages', () => {
      const csv = `email,firstName,lastName,password
invalid-email,Marc,Rovira,Pass12345
clara@encertia.cat,,Vidal,Pass12345
arnau@encertia.cat,Arnau,,Pass12345
eric@encertia.cat,Eric,Zapater,short`

      const result = parseUsersCsv(csv)

      expect(result.totalCount).toBe(4)
      expect(result.validCount).toBe(0)
      expect(result.invalidCount).toBe(4)

      // Row 1: invalid email format
      expect(result.rows[0].isValid).toBe(false)
      expect(result.rows[0].validationErrors.some((e) => e.includes('Format de correu'))).toBe(true)

      // Row 2: missing firstName
      expect(result.rows[1].isValid).toBe(false)
      expect(result.rows[1].validationErrors.some((e) => e.includes('nom és obligatori'))).toBe(true)

      // Row 3: missing lastName
      expect(result.rows[2].isValid).toBe(false)
      expect(result.rows[2].validationErrors.some((e) => e.includes('cognoms són obligatoris'))).toBe(true)

      // Row 4: short password
      expect(result.rows[3].isValid).toBe(false)
      expect(result.rows[3].validationErrors.some((e) => e.includes('mínim 8 caràcters'))).toBe(true)
    })

    it('handles mixed valid and invalid rows', () => {
      const csv = `email,firstName,lastName
valid@encertia.cat,Bona,Fila
no-email,Fila,Dolenta
altra.valida@encertia.cat,Altra,Bona`

      const result = parseUsersCsv(csv)

      expect(result.totalCount).toBe(3)
      expect(result.validCount).toBe(2)
      expect(result.invalidCount).toBe(1)
      expect(result.validItems).toHaveLength(2)
      expect(result.validItems[0].email).toBe('valid@encertia.cat')
      expect(result.validItems[1].email).toBe('altra.valida@encertia.cat')
    })
  })
})
