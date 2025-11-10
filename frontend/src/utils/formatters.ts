export const formatPrice = (priceInCents: number): string => {
  return (priceInCents / 100).toFixed(2);
};

export const parsePrice = (price: string): number => {
  return Math.round(parseFloat(price) * 100);
};

export const isValidPrice = (price: string): boolean => {
  if (!price || price.trim().length === 0) return false;
  const priceInCents = parseFloat(price);
  return !isNaN(priceInCents) && priceInCents >= 0;
};

export const isValidId = (id: string): boolean => {
  if (!id || id.trim().length === 0) return false;
  const numId = parseInt(id, 10);
  return !isNaN(numId) && numId >= 0;
};
