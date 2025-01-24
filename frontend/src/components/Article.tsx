import { ArticleRouteProps } from '../utils/types';
import Header from './Header';

const Article = ({ article }: ArticleRouteProps) => (
	<div>
		<Header />
		<div className='max-w-4xl mx-auto pb-4'>
			<h1 className='text-4xl text-gray-800 mb-4 text-center'>
				{article.title}
			</h1>
			<div
				className='prose text-gray-700'
				dangerouslySetInnerHTML={{ __html: article.content }}
			/>
		</div>
	</div>
);

export default Article;
