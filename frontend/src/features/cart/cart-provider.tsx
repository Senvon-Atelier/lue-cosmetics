import { createContext, useContext, useState, useEffect, useRef, ReactNode } from 'react';
import { useAuth } from '../../lib/auth/auth-provider';
import {
  getCart,
  postCartItems,
  patchCartItemsId,
  deleteCartItemsId,
  postCartMerge,
} from '../../lib/api/generated/rueCosmeticsAPI';
import type { InternalCartCartItemResponse } from '../../lib/api/generated/rueCosmeticsAPI';

// Cart types
interface CartState {
  items: InternalCartCartItemResponse[];
  subtotalGhsMinor: number;
  shippingCostGhsMinor: number;
  totalGhsMinor: number;
  guestToken: string | null;
  isLoading: boolean;
}

interface CartContextType extends CartState {
  addItem: (productId: string, qty: number, name?: string) => Promise<void>;
  updateItem: (itemId: string, qty: number) => Promise<void>;
  removeItem: (itemId: string) => Promise<void>;
  refreshCart: () => Promise<void>;
  itemCount: number;
  wishlistCount: number;
  addToWishlist: () => void;
  removeFromWishlist: () => void;
  // Drawer state
  isDrawerOpen: boolean;
  openDrawer: () => void;
  closeDrawer: () => void;
  // Toast state
  lastAdded: { name: string } | null;
  dismissToast: () => void;
}

const CartContext = createContext<CartContextType | undefined>(undefined);

// Local storage key for guest cart token
const GUEST_CART_KEY = 'rue.guest_cart';

// Cart provider component
export function CartProvider({ children }: { children: ReactNode }) {
  const { isAuthenticated } = useAuth();
  const [state, setState] = useState<CartState>({
    items: [],
    subtotalGhsMinor: 0,
    shippingCostGhsMinor: 0,
    totalGhsMinor: 0,
    guestToken: null,
    isLoading: false,
  });

  // Wishlist state
  const [wishlistCount, setWishlistCount] = useState(0);

  // Drawer state
  const [isDrawerOpen, setIsDrawerOpen] = useState(false);
  const openDrawer = () => setIsDrawerOpen(true);
  const closeDrawer = () => setIsDrawerOpen(false);

  // Toast state
  const [lastAdded, setLastAdded] = useState<{ name: string } | null>(null);
  const toastTimerRef = useRef<number | null>(null);
  const dismissToast = () => {
    if (toastTimerRef.current !== null) clearTimeout(toastTimerRef.current);
    setLastAdded(null);
  };

  // Refresh cart from backend
  const refreshCart = async () => {
    setState((prev) => ({ ...prev, isLoading: true }));
    try {
      const cartData = await getCart();

      // Store guest token if present
      if (cartData?.guest_token) {
        localStorage.setItem(GUEST_CART_KEY, cartData.guest_token);
      }

      setState({
        items: cartData?.items || [],
        subtotalGhsMinor: cartData?.subtotal_ghs_minor || 0,
        shippingCostGhsMinor: cartData?.shipping_cost_ghs_minor || 0,
        totalGhsMinor: cartData?.total_ghs_minor || 0,
        guestToken: cartData?.guest_token || null,
        isLoading: false,
      });
    } catch (error) {
      console.error('Failed to load cart:', error);
      setState((prev) => ({ ...prev, isLoading: false }));
    }
  };

  // Load cart on mount and when auth state changes
  useEffect(() => {
    refreshCart();
  }, [isAuthenticated]);

  // Merge guest cart on login
  useEffect(() => {
    if (isAuthenticated) {
      const guestToken = localStorage.getItem(GUEST_CART_KEY);
      if (guestToken) {
        postCartMerge({ guest_token: guestToken })
          .then(() => {
            localStorage.removeItem(GUEST_CART_KEY);
            return refreshCart();
          })
          .catch((error: unknown) => {
            console.error('Failed to merge guest cart:', error);
          });
      }
    }
  }, [isAuthenticated]);

  // Add item to cart
  const addItem = async (productId: string, qty: number, name?: string) => {
    await postCartItems({ product_id: productId, qty });
    await refreshCart();
    if (toastTimerRef.current !== null) clearTimeout(toastTimerRef.current);
    setLastAdded(name ? { name } : { name: 'Item' });
    toastTimerRef.current = window.setTimeout(() => setLastAdded(null), 2400);
  };

  // Update item quantity
  const updateItem = async (itemId: string, qty: number) => {
    await patchCartItemsId(itemId, { qty });
    await refreshCart();
  };

  // Remove item from cart
  const removeItem = async (itemId: string) => {
    await deleteCartItemsId(itemId);
    await refreshCart();
  };

  // Wishlist helpers
  const addToWishlist = () => setWishlistCount((prev) => prev + 1);
  const removeFromWishlist = () => setWishlistCount((prev) => Math.max(0, prev - 1));

  const itemCount = (state.items || []).reduce((sum, item) => sum + (item.qty || 0), 0);

  return (
    <CartContext.Provider
      value={{
        ...state,
        addItem,
        updateItem,
        removeItem,
        refreshCart,
        itemCount,
        wishlistCount,
        addToWishlist,
        removeFromWishlist,
        isDrawerOpen,
        openDrawer,
        closeDrawer,
        lastAdded,
        dismissToast,
      }}
    >
      {children}
    </CartContext.Provider>
  );
}

// Hook to use cart context
export function useCart() {
  const context = useContext(CartContext);
  if (context === undefined) {
    throw new Error('useCart must be used within a CartProvider');
  }
  return context;
}
