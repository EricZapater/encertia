/**
 * Utilitats de format i renderització de text per a preguntes d'Encertia
 */

/**
 * Detecta si el text conté una etiqueta d'imatge inline com [imatge], [foto] o [image]
 */
export function hasInlineImageTag(text: string | null | undefined): boolean {
  if (!text) return false
  return /\[(?:imatge|foto|image)\]/i.test(text)
}

/**
 * Escapa caràcters HTML perillosos per evitar XSS
 */
export function escapeHtml(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;')
}

/**
 * Converteix text enriquit segur (Markdown bàsic + tags de format + [imatge]) a HTML segur.
 */
export function formatQuestionText(rawText: string | null | undefined, imageUrl?: string | null): string {
  if (!rawText) return ''

  // 1. Escapem qualsevol HTML perillós
  let html = escapeHtml(rawText)

  // 2. Headings (# H1, ## H2, ### H3 a principi de línia amb flag multiline)
  html = html.replace(/^###\s+([^\r\n]+)$/gm, '<h3 class="formatted-h3">$1</h3>')
  html = html.replace(/^##\s+([^\r\n]+)$/gm, '<h2 class="formatted-h2">$1</h2>')
  html = html.replace(/^#\s+([^\r\n]+)$/gm, '<h1 class="formatted-h1">$1</h1>')

  // 3. Negreta: **text** o __text__
  html = html.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
  html = html.replace(/__([^_]+)__/g, '<strong>$1</strong>')

  // 4. Cursiva: *text* o _text_
  html = html.replace(/(^|[^*])\*([^*\n]+)\*([^*]|$)/g, '$1<em>$2</em>$3')
  html = html.replace(/(^|[^_])_([^_\n]+)_([^_]|$)/g, '$1<em>$2</em>$3')

  // 5. Subratllat: <u>text</u> o [u]text[/u]
  html = html.replace(/&lt;u&gt;(.*?)&lt;\/u&gt;/gi, '<u>$1</u>')
  html = html.replace(/\[u\](.*?)\[\/u\]/gi, '<u>$1</u>')

  // 6. Imatge inline: [imatge], [foto], [image]
  const imageRegex = /\[(?:imatge|foto|image)\]/gi
  if (imageUrl) {
    const safeUrl = escapeHtml(imageUrl)
    const imgHtml = `<span class="inline-question-img-wrapper"><img src="${safeUrl}" alt="Imatge de la pregunta" class="inline-question-img" /></span>`
    html = html.replace(imageRegex, imgHtml)
  } else {
    html = html.replace(imageRegex, '')
  }

  // 7. Salts de línia \n -> <br>
  html = html.replace(/\n/g, '<br>')
  html = html.replace(/(<\/h[1-3]>)<br>/g, '$1')

  return html
}
