export function toTitleCase(str: string): string {
  return str
    .split(' ')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ');
}


export async function sleep(ms: number) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

export function arraysEqual<T>(a: T[], b: T[]): boolean {
  const set1 = new Set(a);
  const set2 = new Set(b);

  if (set1.size !== set2.size) return false;
  return [...set1].every(item => set2.has(item));
}
