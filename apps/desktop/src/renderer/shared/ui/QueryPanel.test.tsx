import { render, screen } from '@testing-library/react';
import { QueryPanel } from '@/shared/ui/QueryPanel';

describe('QueryPanel', () => {
  it('renders loading state', () => {
    render(
      <QueryPanel loading={true} error="" title="Test Panel">
        <div data-testid="content">Content</div>
      </QueryPanel>,
    );
    expect(screen.getByText('加载中')).toBeInTheDocument();
  });

  it('renders error state', () => {
    render(
      <QueryPanel loading={false} error="Something went wrong" title="Test Panel">
        <div data-testid="content">Content</div>
      </QueryPanel>,
    );
    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
  });

  it('renders children when ready', () => {
    render(
      <QueryPanel loading={false} error="" title="Test Panel">
        <div data-testid="content">Hello World</div>
      </QueryPanel>,
    );
    expect(screen.getByTestId('content')).toHaveTextContent('Hello World');
  });
});
