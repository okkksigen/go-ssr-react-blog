import { IndexRouteProps } from '../generated';
import ArticlePreview from '../components/ArticlePreview';
import Header from '../components/Header';

const Home = ({ articles }: IndexRouteProps) => (
	<div>
		<Header />
		<div className='max-w-3xl mx-auto'>
			{articles.map((article, i) => (
				<ArticlePreview key={i} article={article} />
			))}
		</div>
	</div>
);

export default Home;
