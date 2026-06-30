import { useState, useEffect } from 'react';
import { useNavigate } from '@tanstack/react-router';
// TODO: Uncomment when wishlist backend is implemented
// import {
//   getWishlist,
//   deleteWishlistProductProductId,
// } from '../../lib/api/generated/rueCosmeticsAPI';
import { Button } from '../shared/ui/button';

type WishlistItem = {
  id: string;
  product_id: string;
  product_name: string;
  product_slug: string;
  product_price: number;
  product_image: string;
  product_brand: string;
  created_at: string;
};

export function AccountWishlist() {
  const navigate = useNavigate();
  const [wishlistItems, setWishlistItems] = useState<WishlistItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadWishlist = async () => {
    setIsLoading(true);
    setError(null);
    try {
      // TODO: Uncomment when wishlist backend is implemented
      // const response = await getWishlist();
      // setWishlistItems(response);
      setWishlistItems([]);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load wishlist');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadWishlist();
  }, []);

  const handleRemoveFromWishlist = async (_productId: string) => {
    try {
      // TODO: Uncomment when wishlist backend is implemented
      // await deleteWishlistProductProductId({
      //   path: { product_id: productId },
      // });
      await loadWishlist();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to remove from wishlist');
    }
  };

  const handleAddToCart = async (_productId: string, productName: string) => {
    // This would require cart context integration
    alert(`Add ${productName} to cart functionality would integrate with cart context`);
  };

  const formatCurrency = (amount: number) => {
    return `GH₵${amount.toFixed(2)}`;
  };

  if (isLoading) {
    return (
      <div className="text-center py-12">
        <div className="text-ink-muted">Loading wishlist...</div>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6">
        <h2 className="font-display text-xl mb-2">Wishlist</h2>
        <p className="text-ink-muted">Items you've saved for later.</p>
      </div>

      {/* Error Message */}
      {error && (
        <div className="bg-rose-50 border border-rose-200 text-rose-800 px-4 py-3 rounded-lg mb-4">
          {error}
        </div>
      )}

      {/* Empty State */}
      {wishlistItems.length === 0 ? (
        <div className="text-center py-12">
          <div className="text-4xl mb-4">❤️</div>
          <h3 className="font-display text-xl mb-2">Your wishlist is empty</h3>
          <p className="text-ink-muted mb-6">Save items you love so you can find them easily later.</p>
          <Button variant="primary" onClick={() => navigate({ to: '/shop' })}>
            Explore Products
          </Button>
        </div>
      ) : (
        /* Wishlist Grid */
        <div className="grid md:grid-cols-3 gap-6">
          {wishlistItems.map((item) => (
            <div
              key={item.id}
              className="bg-white rounded-lg overflow-hidden"
              style={{ border: '1px solid var(--line)' }}
            >
              {/* Product Image */}
              <div className="aspect-square bg-lavender-50 relative">
                {item.product_image ? (
                  <img
                    src={item.product_image}
                    alt={item.product_name}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center text-ink-muted text-xs">
                    No image
                  </div>
                )}
                <div className="absolute top-2 right-2">
                  <button
                    onClick={() => handleRemoveFromWishlist(item.product_id)}
                    className="w-8 h-8 rounded-full bg-white shadow flex items-center justify-center hover:bg-rose-50 transition-colors"
                    title="Remove from wishlist"
                  >
                    ✕
                  </button>
                </div>
              </div>

              {/* Product Info */}
              <div className="p-4">
                <p className="text-xs text-ink-muted mb-1">{item.product_brand}</p>
                <h3 className="font-label font-medium mb-2 line-clamp-2">
                  <a
                    href={`/shop/${item.product_slug}`}
                    className="text-ink hover:text-lavender-700 transition-colors"
                  >
                    {item.product_name}
                  </a>
                </h3>
                <div className="flex items-center justify-between">
                  <div className="font-display font-semibold">
                    {formatCurrency(item.product_price)}
                  </div>
                  <Button
                    variant="primary"
                    size="sm"
                    onClick={() => handleAddToCart(item.product_id, item.product_name)}
                  >
                    Add to Cart
                  </Button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
