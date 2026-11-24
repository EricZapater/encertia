import { describe, it, expect } from 'vitest'
import { formatQuestionText, hasInlineImageTag, escapeHtml } from '../textFormat'

describe('textFormat utilities', () => {
  it('detecta etiquetes d\'imatge inline', () => {
    expect(hasInlineImageTag('Pregunta sense imatge')).toBe(false)
    expect(hasInlineImageTag('Pregunta amb [imatge] inline')).toBe(true)
    expect(hasInlineImageTag('Pregunta amb [IMATGE] majúscules')).toBe(true)
    expect(hasInlineImageTag('Pregunta amb [foto]')).toBe(true)
    expect(hasInlineImageTag('Pregunta amb [image]')).toBe(true)
    expect(hasInlineImageTag(null)).toBe(false)
    expect(hasInlineImageTag('')).toBe(false)
  })

  it('escapa caràcters HTML perillosos per evitar XSS', () => {
    const malicious = '<script>alert("xss")</script> & <b>test</b>'
    const escaped = escapeHtml(malicious)
    expect(escaped).not.toContain('<script>')
    expect(escaped).toContain('&lt;script&gt;')
    expect(escaped).toContain('&amp;')
  })

  it('aplica format de negreta, cursiva i subratllat correctament', () => {
    const raw = 'Text amb **negreta** i *cursiva* i <u>subratllat</u>'
    const formatted = formatQuestionText(raw)
    expect(formatted).toContain('<strong>negreta</strong>')
    expect(formatted).toContain('<em>cursiva</em>')
    expect(formatted).toContain('<u>subratllat</u>')
  })

  it('aplica títols H1, H2 i H3', () => {
    const raw = '# Títol Principal\n## Subtítol\nText normal'
    const formatted = formatQuestionText(raw)
    expect(formatted).toContain('<h1 class="formatted-h1">Títol Principal</h1>')
    expect(formatted).toContain('<h2 class="formatted-h2">Subtítol</h2>')
    expect(formatted).toContain('Text normal')
  })

  it('insereix la imatge al lloc exacte del tag [imatge] si hi ha imageUrl', () => {
    const raw = 'Com es diu aquest monument: [imatge] a Barcelona?'
    const imageUrl = 'https://cdn.encertia.cat/sagrada_familia.jpg'
    const formatted = formatQuestionText(raw, imageUrl)

    expect(formatted).toContain('<span class="inline-question-img-wrapper">')
    expect(formatted).toContain('src="https://cdn.encertia.cat/sagrada_familia.jpg"')
    expect(formatted).toContain('Com es diu aquest monument: <span class="inline-question-img-wrapper">')
    expect(formatted).toContain('</span> a Barcelona?')
  })

  it('elimina el tag [imatge] netament si no hi ha imageUrl', () => {
    const raw = 'Pregunta amb [imatge] però sense foto real'
    const formatted = formatQuestionText(raw, null)
    expect(formatted).not.toContain('[imatge]')
    expect(formatted).not.toContain('<img')
  })
})
