import React from 'react';

export default class ControlPanel extends React.Component {

      render() {
      	       return (<div>
	       <input type="number" defaultValue="1" min="1" max="20" />
	       <input type="button" value="Start" />
	       </div>);
      }

} 