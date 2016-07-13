import React from 'react';
import ReactDOM from 'react-dom';

import ControlPanel from './ControlPanel.jsx';
import Console from './Console.jsx';

class Root extends React.Component {

      render() {
      	       return (<div id="mainDiv">
			<label id="titleLabel">Raft - Leader election simulator</label>
			<ControlPanel />
			<Console />
		      </div>);
	       
      }
}

ReactDOM.render(<Root/>, document.getElementById('root'));

