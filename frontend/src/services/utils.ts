export function toTitleCase(str: string): string {
  return str
    .split(' ')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ');
}


export async function sleep(ms: number) {
  return new Promise(resolve => setTimeout(resolve, ms));
}
