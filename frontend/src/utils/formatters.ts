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

export const createIdInputHandler = (setter: (value: string) => void) => {
  return (e: Event) => {
    const target = e.target as HTMLInputElement;
    const value = target.value;
    if (value === "" || /^\d+$/.test(value)) {
      setter(value);
    }
  };
};

export const createInputHandler = (setter: (value: string) => void) => {
  return (e: Event) => {
    const target = e.target as HTMLInputElement;
    setter(target.value);
  };
};

export const createSelectHandler = (setter: (value: string) => void) => {
  return (e: Event) => {
    const target = e.target as HTMLSelectElement;
    setter(target.value);
  };
};

export const PROMO_CARD_STYLE = {
  backgroundColor: "rgba(100, 200, 100, 0.05)",
  borderColor: "rgba(100, 200, 100, 0.2)",
};
