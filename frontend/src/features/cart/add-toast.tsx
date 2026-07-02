import { Icon } from '../shared/ui/icons';

interface AddToastProps {
  lastAdded: { name: string } | null;
  onView: () => void;
  onDismiss: () => void;
}

export function AddToast({ lastAdded, onView, onDismiss }: AddToastProps) {
  if (!lastAdded) return null;
  return (
    <div className="toast" role="status" aria-live="polite">
      <Icon name="check" size={14} />
      <span>
        <strong>Added.</strong> {lastAdded.name}
      </span>
      <button
        onClick={() => {
          onDismiss();
          onView();
        }}
      >
        View bag
      </button>
    </div>
  );
}
