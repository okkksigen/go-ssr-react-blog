import Article from '../components/Article';
import { ArticleRouteProps } from '../utils/types';

const ArticlePage = ({ article }: ArticleRouteProps) => (
	<Article article={article} />
);

export default ArticlePage;
