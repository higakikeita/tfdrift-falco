/**
 * Tests for ProviderIcon component
 * Tests AWS, GCP, and Azure icon rendering with various size and style props
 */

import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/react';
import { ProviderIcon } from './ProviderIcons';

describe('ProviderIcon', () => {
  describe('AWS Icon', () => {
    it('renders AWS icon', () => {
      const { container } = render(<ProviderIcon provider="aws" />);
      const svg = container.querySelector('svg');

      expect(svg).toBeInTheDocument();
    });

    it('renders AWS icon with default size', () => {
      const { container } = render(<ProviderIcon provider="aws" />);
      const svg = container.querySelector('svg');

      expect(svg).toHaveAttribute('width', '24');
      expect(svg).toHaveAttribute('height', '24');
    });

    it('renders AWS icon with custom size', () => {
      const { container } = render(<ProviderIcon provider="aws" size={48} />);
      const svg = container.querySelector('svg');

      expect(svg).toHaveAttribute('width', '48');
      expect(svg).toHaveAttribute('height', '48');
    });

    it('renders AWS icon with small size', () => {
      const { container } = render(<ProviderIcon provider="aws" size={16} />);
      const svg = container.querySelector('svg');

      expect(svg).toHaveAttribute('width', '16');
      expect(svg).toHaveAttribute('height', '16');
    });

    it('renders AWS icon with large size', () => {
      const { container } = render(<ProviderIcon provider="aws" size={96} />);
      const svg = container.querySelector('svg');

      expect(svg).toHaveAttribute('width', '96');
      expect(svg).toHaveAttribute('height', '96');
    });

    it('applies custom className to AWS icon', () => {
      const { container } = render(
        <ProviderIcon provider="aws" className="custom-class" />
      );
      const svg = container.querySelector('svg');

      expect(svg).toHaveClass('custom-class');
    });

    it('renders AWS paths with orange color', () => {
      const { container } = render(<ProviderIcon provider="aws" />);
      const paths = container.querySelectorAll('path');

      expect(paths.length).toBeGreaterThan(0);
      paths.forEach((path) => {
        const fill = path.getAttribute('fill');
        expect(fill).toMatch(/#FF9900|#ff9900/);
      });
    });

    it('AWS icon viewBox is correct', () => {
      const { container } = render(<ProviderIcon provider="aws" />);
      const svg = container.querySelector('svg');

      expect(svg).toHaveAttribute('viewBox', '0 0 24 24');
    });

    it('AWS icon is SVG element', () => {
      const { container } = render(<ProviderIcon provider="aws" />);
      const svg = container.querySelector('svg');

      expect(svg?.tagName).toBe('svg');
    });

    it('renders AWS icon without xmlns attribute if not present', () => {
      const { container } = render(<ProviderIcon provider="aws" />);
      const svg = container.querySelector('svg');

      const xmlns = svg?.getAttributeNS('http://www.w3.org/2000/xmlns/', 'xmlns');
      expect(xmlns).toBeDefined();
    });
  });

  describe('GCP Icon', () => {
    it('renders GCP icon', () => {
      const { container } = render(<ProviderIcon provider="gcp" />);
      const svg = container.querySelector('svg');

      expect(svg).toBeInTheDocument();
    });

    it('renders GCP icon with default size', () => {
      const { container } = render(<ProviderIcon provider="gcp" />);
      const svg = container.querySelector('svg');

      expect(svg).toHaveAttribute('width', '24');
      expect(svg).toHaveAttribute('height', '24');
    });

    it('renders GCP icon with custom size', () => {
      const { container } = render(<ProviderIcon provider="gcp" size={40} />);
      const svg = container.querySelector('svg');

      expect(svg).toHaveAttribute('width', '40');
      expect(svg).toHaveAttribute('height', '40');
    });

    it('applies custom className to GCP icon', () => {
      const { container } = render(
        <ProviderIcon provider="gcp" className="gcp-custom" />
      );
      const svg = container.querySelector('svg');

      expect(svg).toHaveClass('gcp-custom');
    });

    it('renders GCP logo with multiple colors', () => {
      const { container } = render(<ProviderIcon provider="gcp" />);
      const paths = container.querySelectorAll('path');

      const colors = new Set();
      paths.forEach((path) => {
        const fill = path.getAttribute('fill');
        if (fill && fill.startsWith('#')) {
          colors.add(fill);
        }
      });

      expect(colors.size).toBeGreaterThan(1);
    });

    it('GCP icon viewBox is correct', () => {
      const { container } = render(<ProviderIcon provider="gcp" />);
      const svg = container.querySelector('svg');

      expect(svg).toHaveAttribute('viewBox', '0 0 24 24');
    });

    it('GCP icon contains radial gradient definition', () => {
      const { container } = render(<ProviderIcon provider="gcp" />);
      const defs = container.querySelector('defs');

      expect(defs).toBeInTheDocument();
    });
  });

  describe('Azure Icon', () => {
    it('renders Azure icon', () => {
      const { container } = render(<ProviderIcon provider="azure" />);
      const svg = container.querySelector('svg');

      expect(svg).toBeInTheDocument();
    });

    it('renders Azure icon with default size', () => {
      const { container } = render(<ProviderIcon provider="azure" />);
      const svg = container.querySelector('svg');

      expect(svg).toHaveAttribute('width', '24');
      expect(svg).toHaveAttribute('height', '24');
    });

    it('renders Azure icon with custom size', () => {
      const { container } = render(<ProviderIcon provider="azure" size={32} />);
      const svg = container.querySelector('svg');

      expect(svg).toHaveAttribute('width', '32');
      expect(svg).toHaveAttribute('height', '32');
    });

    it('applies custom className to Azure icon', () => {
      const { container } = render(
        <ProviderIcon provider="azure" className="azure-custom" />
      );
      const svg = container.querySelector('svg');

      expect(svg).toHaveClass('azure-custom');
    });

    it('renders Azure logo with blue color', () => {
      const { container } = render(<ProviderIcon provider="azure" />);
      const paths = container.querySelectorAll('path');

      expect(paths.length).toBeGreaterThan(0);
      paths.forEach((path) => {
        const fill = path.getAttribute('fill');
        if (fill && fill.startsWith('#')) {
          expect(fill).toMatch(/#0078D4|#0078d4|url/i);
        }
      });
    });

    it('Azure icon viewBox is correct', () => {
      const { container } = render(<ProviderIcon provider="azure" />);
      const svg = container.querySelector('svg');

      expect(svg).toHaveAttribute('viewBox', '0 0 24 24');
    });

    it('Azure icon contains linear gradient definition', () => {
      const { container } = render(<ProviderIcon provider="azure" />);
      const defs = container.querySelector('defs');

      expect(defs).toBeInTheDocument();
    });
  });

  describe('Size Variations', () => {
    it('handles very small size', () => {
      const { container } = render(<ProviderIcon provider="aws" size={8} />);
      const svg = container.querySelector('svg');

      expect(svg).toHaveAttribute('width', '8');
      expect(svg).toHaveAttribute('height', '8');
    });

    it('handles very large size', () => {
      const { container } = render(<ProviderIcon provider="gcp" size={256} />);
      const svg = container.querySelector('svg');

      expect(svg).toHaveAttribute('width', '256');
      expect(svg).toHaveAttribute('height', '256');
    });

    it('handles fractional sizes', () => {
      const { container } = render(<ProviderIcon provider="azure" size={24.5} />);
      const svg = container.querySelector('svg');

      expect(svg).toHaveAttribute('width', '24.5');
      expect(svg).toHaveAttribute('height', '24.5');
    });

    it('maintains aspect ratio for all sizes', () => {
      [16, 24, 32, 48, 64, 96].forEach((size) => {
        const { container } = render(<ProviderIcon provider="aws" size={size} />);
        const svg = container.querySelector('svg');

        expect(svg).toHaveAttribute('width', String(size));
        expect(svg).toHaveAttribute('height', String(size));
      });
    });
  });

  describe('Style and Class Props', () => {
    it('combines multiple classNames', () => {
      const { container } = render(
        <ProviderIcon
          provider="aws"
          className="icon-class additional-class"
        />
      );
      const svg = container.querySelector('svg');

      expect(svg).toHaveClass('icon-class');
      expect(svg).toHaveClass('additional-class');
    });

    it('handles empty className', () => {
      const { container } = render(
        <ProviderIcon provider="gcp" className="" />
      );
      const svg = container.querySelector('svg');

      expect(svg).toBeInTheDocument();
    });

    it('applies all three providers with classes', () => {
      const providers = ['aws', 'gcp', 'azure'] as const;
      const className = 'provider-icon';

      providers.forEach((provider) => {
        const { container } = render(
          <ProviderIcon provider={provider} className={className} />
        );
        const svg = container.querySelector('svg');
        expect(svg).toHaveClass(className);
      });
    });
  });

  describe('SVG Structure', () => {
    it('AWS icon has proper SVG namespace', () => {
      const { container } = render(<ProviderIcon provider="aws" />);
      const svg = container.querySelector('svg');

      expect(svg?.namespaceURI).toBe('http://www.w3.org/2000/svg');
    });

    it('GCP icon has proper SVG namespace', () => {
      const { container } = render(<ProviderIcon provider="gcp" />);
      const svg = container.querySelector('svg');

      expect(svg?.namespaceURI).toBe('http://www.w3.org/2000/svg');
    });

    it('Azure icon has proper SVG namespace', () => {
      const { container } = render(<ProviderIcon provider="azure" />);
      const svg = container.querySelector('svg');

      expect(svg?.namespaceURI).toBe('http://www.w3.org/2000/svg');
    });

    it('AWS icon fill attribute is set to none', () => {
      const { container } = render(<ProviderIcon provider="aws" />);
      const svg = container.querySelector('svg');

      expect(svg).toHaveAttribute('fill', 'none');
    });

    it('GCP icon fill attribute is set to none', () => {
      const { container } = render(<ProviderIcon provider="gcp" />);
      const svg = container.querySelector('svg');

      expect(svg).toHaveAttribute('fill', 'none');
    });

    it('Azure icon fill attribute is set to none', () => {
      const { container } = render(<ProviderIcon provider="azure" />);
      const svg = container.querySelector('svg');

      expect(svg).toHaveAttribute('fill', 'none');
    });
  });

  describe('Rendering Multiple Icons', () => {
    it('renders multiple instances without errors', () => {
      const { container } = render(
        <>
          <ProviderIcon provider="aws" size={24} />
          <ProviderIcon provider="gcp" size={24} />
          <ProviderIcon provider="azure" size={24} />
        </>
      );

      const svgs = container.querySelectorAll('svg');
      expect(svgs).toHaveLength(3);
    });

    it('renders multiple instances with different sizes', () => {
      const { container } = render(
        <>
          <ProviderIcon provider="aws" size={16} />
          <ProviderIcon provider="aws" size={32} />
          <ProviderIcon provider="aws" size={64} />
        </>
      );

      const svgs = container.querySelectorAll('svg');
      expect(svgs[0]).toHaveAttribute('width', '16');
      expect(svgs[1]).toHaveAttribute('width', '32');
      expect(svgs[2]).toHaveAttribute('width', '64');
    });

    it('renders multiple instances with different classes', () => {
      const { container } = render(
        <>
          <ProviderIcon provider="aws" className="aws-icon" />
          <ProviderIcon provider="gcp" className="gcp-icon" />
          <ProviderIcon provider="azure" className="azure-icon" />
        </>
      );

      const svgs = container.querySelectorAll('svg');
      expect(svgs[0]).toHaveClass('aws-icon');
      expect(svgs[1]).toHaveClass('gcp-icon');
      expect(svgs[2]).toHaveClass('azure-icon');
    });
  });

  describe('Default Behavior', () => {
    it('uses default size when not provided', () => {
      const { container } = render(<ProviderIcon provider="aws" />);
      const svg = container.querySelector('svg');

      expect(svg).toHaveAttribute('width', '24');
      expect(svg).toHaveAttribute('height', '24');
    });

    it('uses empty className when not provided', () => {
      const { container } = render(<ProviderIcon provider="gcp" />);
      const svg = container.querySelector('svg');

      expect(svg).toHaveAttribute('class', '');
    });

    it('renders all three providers with defaults', () => {
      const providers = ['aws', 'gcp', 'azure'] as const;

      providers.forEach((provider) => {
        const { container } = render(
          <ProviderIcon provider={provider} />
        );
        const svg = container.querySelector('svg');

        expect(svg).toHaveAttribute('width', '24');
        expect(svg).toHaveAttribute('height', '24');
        expect(svg).toHaveAttribute('viewBox', '0 0 24 24');
      });
    });
  });

  describe('Color Consistency', () => {
    it('AWS icon uses consistent orange color', () => {
      const { container } = render(<ProviderIcon provider="aws" />);
      const paths = container.querySelectorAll('path[fill="#FF9900"]');

      expect(paths.length).toBeGreaterThan(0);
    });

    it('GCP icon uses multiple brand colors', () => {
      const { container } = render(<ProviderIcon provider="gcp" />);
      const colorSet = new Set<string>();

      container.querySelectorAll('path').forEach((path) => {
        const fill = path.getAttribute('fill');
        if (fill && fill.startsWith('#')) {
          colorSet.add(fill);
        }
      });

      expect(colorSet.size).toBeGreaterThan(1);
    });

    it('Azure icon uses consistent blue color', () => {
      const { container } = render(<ProviderIcon provider="azure" />);
      const paths = container.querySelectorAll('path[fill="#0078D4"]');

      expect(paths.length).toBeGreaterThan(0);
    });
  });

  describe('Edge Cases', () => {
    it('handles zero size gracefully', () => {
      const { container } = render(<ProviderIcon provider="aws" size={0} />);
      const svg = container.querySelector('svg');

      expect(svg).toHaveAttribute('width', '0');
      expect(svg).toHaveAttribute('height', '0');
    });

    it('handles negative size gracefully', () => {
      const { container } = render(<ProviderIcon provider="gcp" size={-24} />);
      const svg = container.querySelector('svg');

      expect(svg).toHaveAttribute('width', '-24');
      expect(svg).toHaveAttribute('height', '-24');
    });

    it('handles very long className', () => {
      const longClass = 'class1 class2 class3 class4 class5 class6';
      const { container } = render(
        <ProviderIcon provider="azure" className={longClass} />
      );
      const svg = container.querySelector('svg');

      expect(svg).toHaveClass('class1');
      expect(svg).toHaveClass('class6');
    });

    it('handles special characters in className', () => {
      const { container } = render(
        <ProviderIcon provider="aws" className="icon-class_name--special" />
      );
      const svg = container.querySelector('svg');

      expect(svg).toHaveClass('icon-class_name--special');
    });
  });
});
