export function formatDate (date: string, options?: Intl.DateTimeFormatOptions) {
  return new Date(date).toLocaleString('ru', options)
}