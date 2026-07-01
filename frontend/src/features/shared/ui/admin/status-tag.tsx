interface StatusTagProps {
  status: string;
}

export function StatusTag({ status }: StatusTagProps) {
  const getStatusConfig = (s: string) => {
    const lower = s.toLowerCase();
    // Order status
    if (lower === 'paid' || lower === 'delivered' || lower === 'shipped') {
      return {
        label: s.charAt(0).toUpperCase() + s.slice(1),
        className: 'bg-green-100 text-green-700',
      };
    }
    if (lower === 'pending' || lower === 'processing') {
      return {
        label: s.charAt(0).toUpperCase() + s.slice(1),
        className: 'bg-yellow-100 text-yellow-700',
      };
    }
    if (lower === 'cancelled' || lower === 'failed') {
      return {
        label: s.charAt(0).toUpperCase() + s.slice(1),
        className: 'bg-red-100 text-red-700',
      };
    }
    // Product status
    if (lower === 'live') {
      return { label: 'Live', className: 'bg-green-100 text-green-700' };
    }
    if (lower === 'low' || lower === 'low stock') {
      return { label: 'Low stock', className: 'bg-yellow-100 text-yellow-700' };
    }
    if (lower === 'oos' || lower === 'out of stock') {
      return { label: 'Out of stock', className: 'bg-red-100 text-red-700' };
    }
    if (lower === 'draft') {
      return { label: 'Draft', className: 'bg-gray-100 text-gray-700' };
    }
    // Default
    return {
      label: s.charAt(0).toUpperCase() + s.slice(1),
      className: 'bg-gray-100 text-gray-700',
    };
  };

  const config = getStatusConfig(status);

  return (
    <span className={`inline-flex px-2 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider ${config.className}`}>
      {config.label}
    </span>
  );
}
