// @vitest-environment jsdom
import '@testing-library/jest-dom/vitest';
import { describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AddToast } from './add-toast';

describe('AddToast', () => {
  it('renders nothing when there is no recent add', () => {
    const { container } = render(
      <AddToast lastAdded={null} onView={() => {}} onDismiss={() => {}} />,
    );
    expect(container.firstChild).toBeNull();
  });

  it('shows the product name and opens the bag on click', async () => {
    const onView = vi.fn();
    render(<AddToast lastAdded={{ name: 'Rose Serum' }} onView={onView} onDismiss={() => {}} />);
    expect(screen.getByText(/Rose Serum/)).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', { name: /view bag/i }));
    expect(onView).toHaveBeenCalledOnce();
  });
});
