import ArticlePreview from '../components/ArticlePreview';
import Header from '../components/Header';
import { IndexRouteProps } from '../utils/types';

const Home = ({ articles }: IndexRouteProps) => (
	<div>
		<Header />
		<div className='max-w-3xl mx-auto'>
			{articles && articles.length > 0 ? (
				<div>
					{articles.map(article => (
						<ArticlePreview key={article.slug} article={article} />
					))}
				</div>
			) : (
				<p>Пока нет ни одной статьи.</p>
			)}
		</div>
	</div>
);

export default Home;
