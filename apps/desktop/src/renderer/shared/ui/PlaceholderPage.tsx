type PlaceholderPageProps = {
  section: string;
  title: string;
  description: string;
  bullets: string[];
};

export function PlaceholderPage(props: PlaceholderPageProps) {
  const { section, title, description, bullets } = props;

  return (
    <section className="placeholder-page">
      <div className="placeholder-hero">
        <div>
          <p className="placeholder-section">{section}</p>
          <h3 className="placeholder-title">{title}</h3>
        </div>
        <span className="status-pill">Placeholder</span>
      </div>
      <p className="placeholder-description">{description}</p>
      <div className="placeholder-grid">
        {bullets.map((item) => (
          <article key={item} className="placeholder-card">
            <h4>{item}</h4>
            <p>Module contract is defined in docs and can be connected to live DTOs and event streams from this shell.</p>
          </article>
        ))}
      </div>
    </section>
  );
}
