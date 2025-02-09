export function countryISOtoUnicode(isoCode: string) {
  if (isoCode.length > 2 || isoCode.match(/\d/)) return isoCode;
  return isoCode.toUpperCase().replace(/./g, (char) => String.fromCodePoint(char.charCodeAt(0) + 127397));
}
