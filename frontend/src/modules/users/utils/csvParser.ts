/**
 * Utilitats per al parseig i validació de fitxers CSV d'usuaris.
 */
import type { BatchUserItem, UserRole } from '../types'

export interface ParsedCsvRow {
  rowNumber: number
  data: BatchUserItem
  isValid: boolean
  validationErrors: string[]
  rawRow: Record<string, string>
}

export interface CsvParseResult {
  rows: ParsedCsvRow[]
  validCount: number
  invalidCount: number
  totalCount: number
  validItems: BatchUserItem[]
  errors: string[]
}

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

/**
 * Detecta el delimitador més probable (coma, punt i coma, tabulació)
 */
export function detectDelimiter(text: string): string {
  const firstLine = text.split(/\r?\n/)[0] || ''
  const commas = (firstLine.match(/,/g) || []).length
  const semicolons = (firstLine.match(/;/g) || []).length
  const tabs = (firstLine.match(/\t/g) || []).length

  if (semicolons > commas && semicolons > tabs) return ';'
  if (tabs > commas && tabs > semicolons) return '\t'
  return ','
}

/**
 * Divideix una línia CSV tenint en compte cometes dobles
 */
export function splitCsvLine(line: string, delimiter: string): string[] {
  const result: string[] = []
  let current = ''
  let inQuotes = false

  for (let i = 0; i < line.length; i++) {
    const char = line[i]

    if (char === '"') {
      if (inQuotes && line[i + 1] === '"') {
        current += '"'
        i++
      } else {
        inQuotes = !inQuotes
      }
    } else if (char === delimiter && !inQuotes) {
      result.push(current.trim())
      current = ''
    } else {
      current += char
    }
  }
  result.push(current.trim())
  return result
}

/**
 * Normalitza el nom d'una capçalera per reconèixer columnes
 */
function normalizeHeader(header: string): string {
  return header
    .toLowerCase()
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .replace(/[^a-z0-9]/g, '')
}

/**
 * Parseja el text complet d'un fitxer CSV i el valida
 */
export function parseUsersCsv(csvText: string, defaultRole: UserRole = 'student'): CsvParseResult {
  const lines = csvText
    .split(/\r?\n/)
    .map((l) => l.trim())
    .filter((l) => l.length > 0)

  if (lines.length === 0) {
    return {
      rows: [],
      validCount: 0,
      invalidCount: 0,
      totalCount: 0,
      validItems: [],
      errors: ['El fitxer està buit.']
    }
  }

  const delimiter = detectDelimiter(csvText)
  const headerLine = lines[0]
  const rawHeaders = splitCsvLine(headerLine, delimiter)
  const normalizedHeaders = rawHeaders.map(normalizeHeader)

  // Mapes de columnes
  let emailIdx = normalizedHeaders.findIndex((h) =>
    ['email', 'correu', 'correuelectronic', 'mail'].includes(h)
  )
  let firstNameIdx = normalizedHeaders.findIndex((h) =>
    ['firstname', 'nom', 'first', 'name'].includes(h)
  )
  let lastNameIdx = normalizedHeaders.findIndex((h) =>
    ['lastname', 'cognom', 'cognoms', 'surname', 'last'].includes(h)
  )
  let roleIdx = normalizedHeaders.findIndex((h) => ['role', 'rol', 'perfil'].includes(h))
  let passwordIdx = normalizedHeaders.findIndex((h) =>
    ['password', 'contrasenya', 'clau', 'pass'].includes(h)
  )

  // Si no té capçaleres reconegudes, intentem per posició defecte si té 3-5 columnes
  const hasValidHeader = emailIdx !== -1 || firstNameIdx !== -1 || lastNameIdx !== -1
  let startIndex = 1

  if (!hasValidHeader) {
    // Si la primera línia sembla contenir dades directes (ex. un email)
    if (rawHeaders.some((cell) => cell.includes('@'))) {
      startIndex = 0
      emailIdx = 0
      firstNameIdx = 1
      lastNameIdx = 2
      roleIdx = 3
      passwordIdx = 4
    } else {
      return {
        rows: [],
        validCount: 0,
        invalidCount: 0,
        totalCount: 0,
        validItems: [],
        errors: [
          'No s’han trobat capçaleres vàlides. El CSV ha de contenir columnes com: email, nom, cognoms.'
        ]
      }
    }
  }

  const parsedRows: ParsedCsvRow[] = []

  for (let i = startIndex; i < lines.length; i++) {
    const rowNumber = i + 1 // 1-indexed
    const line = lines[i]
    if (!line) continue

    const cells = splitCsvLine(line, delimiter)
    const rawRow: Record<string, string> = {}
    rawHeaders.forEach((h, idx) => {
      rawRow[h] = cells[idx] || ''
    })

    const email = (emailIdx !== -1 && cells[emailIdx] ? cells[emailIdx] : '').trim()
    const firstName = (firstNameIdx !== -1 && cells[firstNameIdx] ? cells[firstNameIdx] : '').trim()
    const lastName = (lastNameIdx !== -1 && cells[lastNameIdx] ? cells[lastNameIdx] : '').trim()
    let rawRole = (roleIdx !== -1 && cells[roleIdx] ? cells[roleIdx] : '').trim().toLowerCase()
    const password = (passwordIdx !== -1 && cells[passwordIdx] ? cells[passwordIdx] : '').trim()

    let role: UserRole = defaultRole
    if (rawRole === 'admin' || rawRole === 'administrador') {
      role = 'admin'
    } else if (rawRole === 'teacher' || rawRole === 'professor' || rawRole === 'docent') {
      role = 'teacher'
    } else if (rawRole === 'student' || rawRole === 'alumne' || rawRole === 'estudiant') {
      role = 'student'
    }

    const validationErrors: string[] = []

    if (!email) {
      validationErrors.push('El correu electrònic és obligatori.')
    } else if (!EMAIL_REGEX.test(email)) {
      validationErrors.push(`Format de correu invàlid: "${email}".`)
    }

    if (!firstName) {
      validationErrors.push('El nom és obligatori.')
    }

    if (!lastName) {
      validationErrors.push('Els cognoms són obligatoris.')
    }

    if (password && password.length < 8) {
      validationErrors.push('La contrasenya especificada ha de tenir com a mínim 8 caràcters.')
    }

    const item: BatchUserItem = {
      email,
      firstName,
      lastName,
      role,
      ...(password ? { password } : {})
    }

    parsedRows.push({
      rowNumber,
      data: item,
      isValid: validationErrors.length === 0,
      validationErrors,
      rawRow
    })
  }

  const validItems = parsedRows.filter((r) => r.isValid).map((r) => r.data)
  const validCount = validItems.length
  const invalidCount = parsedRows.length - validCount

  return {
    rows: parsedRows,
    validCount,
    invalidCount,
    totalCount: parsedRows.length,
    validItems,
    errors: []
  }
}
