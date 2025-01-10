import Article from '../components/Article';
import { ArticleRouteProps } from '../generated';

const ArticlePage = ({ article }: ArticleRouteProps) => (
	<Article article={article} />
);

export default ArticlePage;
