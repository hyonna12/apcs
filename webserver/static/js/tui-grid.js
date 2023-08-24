
/* 
    1. div_id -> div 아이디
    2. hegith -> tui-grid의 높이
    3. headers -> 테이블 헤더 옵션
    4. columns -> 컬럼
    5. data_list -> 데이터 리스트
*/

/* 
    headers: [{ type: 'rowNum', width: 50, align: 'center'}]

    columns: [
        { header: '컬럼_1', name: 'col_1', align: 'center', width: 150, sortable: true, filter: {type: 'text', showApplyBtn: true, showClearBtn: true} },
        { header: '컬럼_2', name: 'col_2', align: 'left' },
        { header: '컬럼_3', name: 'col_3', width: 120, align: 'center' },
        { header: '컬럼_4', name: 'col_4', width: 120, align: 'left'}
    ],

    data : [
        { col_1: 'col_1', col_2: 'col_2', col_3: '2022-02-22', col_4: '11', },
        { col_1: 'col_1', col_2: 'col_2', col_3: '2022-02-23', col_4: '22', },
        { col_1: 'col_1', col_2: 'col_2', col_3: '2022-02-24', col_4: '33', },
    ]; 
*/

function craeteTuiGrid(div_id, height, headers, columns, data_list){

    var Grid = tui.Grid;
    let grid = new Grid({
        el: document.getElementById(div_id),
        scrollX: true,
        scrollY: true,
        minBodyHeight: height,
        bodyHeight: height,
        rowHeaders: headers,
        columns: columns,
        data: data_list
    });
    Grid.setLanguage('ko');
    
    Grid.applyTheme('striped');
    // $('.tui-grid-header-area').css('background-color', '#ffffff')

    return grid;
}


function createReqTuiGrid(div_id, height, headers, columns, api_info, params){

    /* 
        dataSource data : {
            createData: { url: '요청 URL', method: 'POST'}, 
            readData: { url: '요청 URL', method: 'GET'},
            updateData: {url: '요청 URL', method:'PUT'}
        }

        * parameter 받기 (server)
        readData => req.query;
        createData => req.body.createdRows;
        updateData =>  req.body.updatedRows;

        * readData return type
        {
            data: {
                contents : [{...}, {...}, {...}, ...]
            },
            result : true
        }

    */

    var dataSource = {
        withCredentials: false,
        initialRequest: false,
        contentType: 'application/json',
        api: api_info
    }

    var Grid = tui.Grid;
    let grid = new Grid({
        el: document.getElementById(div_id),
        scrollX: true,
        scrollY: true,
        minBodyHeight: height,
        bodyHeight: height,
        rowHeaders: headers,
        columns: columns,
        data: dataSource
    });
    Grid.applyTheme('striped');
    Grid.setLanguage('ko');

    grid.readData(1, params, false);

    return grid;
}


/* 
    사용법
    let createRows = grid.getModifiedRows().createdRows;
    let updateRows = grid.getModifiedRows().updatedRows;
    let isSave = isTuiGridRegist(columns, rows);

    if(isSave){
        tuiGridSave(grid);
    }

*/
function tuiGridSave( grid ){
    // 편집 종료.
    grid.finishEditing();
    grid.blur();

    if(grid.getModifiedRows().createdRows.length > 0 || grid.getModifiedRows().updatedRows.length > 0){
        // 요청 초기값 설정.
        grid.requestCompleteFlag = false;
        grid.requestCount = 0; // 요청 갯수를 파악하는 이유 - 등록하시겠습니까?/수정하시겠습니까? confirm에서 아니오를 누를 경우 요청개수가 달라질수 있어서...
        grid.responseCount = 0;

        // 요청전 요청갯수 설정
        grid.on('beforeRequest', function(ev) {
            grid.requestCount = grid.requestCount + 1;
        });

        // 응답시 응답갯수 설정.
        grid.on('response', function(ev) {
            grid.responseCount = grid.responseCount + 1;
            if( grid.requestCompleteFlag && grid.responseCount >= grid.requestCount ) {
                // 응답완료시 grid readData호출. (요청 초기값 null처리, 이벤트 리스너 삭제)
                grid.requestCount = null;
                grid.responseCount = null;
                grid.requestCompleteFlag = null;
                grid.off('beforeRequest');
                grid.off('response');
                grid.readData(1, '', false);
            }
        });

        return new Promise(function (resolve, reject){
            if(grid.getModifiedRows().createdRows.length > 0){
                grid.request('createData');
            }
            resolve(resolve);
        })
        .then(result => {
            if(grid.getModifiedRows().updatedRows.length > 0){
                grid.request('updateData',{
                    checkedOnly: false
                });
                return result;
            }
        })
        .then(result => {
            grid.requestCompleteFlag = true;
            // grid.readData(1, '', false); //  마지막 updateData응답 보다 readData응답이 빠른 경우 때문에 업데이트 내용이 반영되지 않는 문제 발생.
        })
        .catch(error => {
            console.log('error :: ', error);
            commonErrorHandler(error)
        })
    }
}
/* 
    tui-grid 빈 칸 체크
*/
function isTuiGridRegist( columns, rows){
    let isSave = true;
    for(row of rows){
        let rslt = isTuiGridSave(columns, row);
        if(!rslt){
            isSave = false;
            alert('정보를 모두 입력하세요.');
            return isSave;
        }
    }
    return isSave;
}

function isTuiGridSave( colums, row ){
    let colNames = [];
    for(col of colums){
        colNames.push(col.name);
    }
    for(name of colNames){
        if(row[name] === '' || typeof row[name] === 'undefined'){
            return false;
        }
    }
    return true;
}

const DB_ERROR_CODE = {
    'FOREIGN_KEY_CONSTRAINT_FAIL_DELETE' : {'code': 1451, 'message': 'Cannot delete or update a parent row: a foreign key constraint fails'},
    'FOREIGN_KEY_CONSTRAINT_FAIL_UPDATE' : {'code': 1452, 'message': 'Cannot add or update a child row: a foreign key constraint fails'},
}

/**
 * axios요청시 catch문 error 공통 처리 함수.
 * @param {Object} error - axios요청시 catch문 파라미터 error
 */
function commonErrorHandler(error){
    // console.dir(error);
    if (error.response) {
        console.log('::::::::::::::::::: server error :::::::::::::::::::');
        console.dir(error.response.data);
        if(error.response.data.errno === DB_ERROR_CODE.FOREIGN_KEY_CONSTRAINT_FAIL_DELETE.code) {
            // Cannot delete or update a parent row: a foreign key constraint fails
            alert(`삭제 실패! \n[원인 - 하위 테이블에서 참조하는 데이터 존재]`);
        } else if(error.response.data.errno === DB_ERROR_CODE.FOREIGN_KEY_CONSTRAINT_FAIL_UPDATE.code) {
            // Cannot delete or update a parent row: a foreign key constraint fails
            alert(`저장 실패! \n[원인 - 상위 테이블 참조 데이터 미존재]`);
        }
    } else {
        // Something happened in setting up the request that triggered an Error
        console.log('Error', error.message);
    }
}



function checkboxFunc(grid, rowKey){
	const label = document.createElement('label');
	label.className = 'checkbox';
	label.setAttribute('for', String(rowKey));

	const hiddenInput = document.createElement('input');
	hiddenInput.className = 'hidden-input';
	hiddenInput.id = String(rowKey);

	const customInput = document.createElement('span');
	customInput.className = 'custom-input';

	label.appendChild(hiddenInput);
	label.appendChild(customInput);

	hiddenInput.type = 'checkbox';
	hiddenInput.addEventListener('change', () => {
		if (hiddenInput.checked) {
			grid.check(rowKey);
		} else {
			grid.uncheck(rowKey);
		}
	});

	return label;
}


// tui-grid checkbox 기능 추가 class
class CheckboxRenderer {
	constructor(props) {
        const { grid, rowKey } = props;
        if(grid.getModifiedRows().createdRows.length > 0){
            let arr = [];
            let rows = grid.getModifiedRows().createdRows;
            for(let i=0; i<rows.length; i++){
                arr.push(rows[i].rowKey);
            }

            if(arr.includes(rowKey)){
                const el = document.createElement('button');
                el.id = String(rowKey);

                el.classList.add('deletebtn');
                this.el = el;
                this.render(props);
            }else{
                this.el = checkboxFunc(grid, rowKey);
                this.render(props);
            }
        }else{
            this.el = checkboxFunc(grid, rowKey);
            this.render(props);
        }
	}

	getElement() {
        return this.el;
	}

	render(props) {
		const { grid } = props;
		if(grid.getModifiedRows().createdRows.length === 0){
			const hiddenInput = this.el.querySelector('.hidden-input');
			const checked = Boolean(props.value);

			hiddenInput.checked = checked;
		}
		else{
			let checkboxes = $('.hidden-input');
			for(var i=0; i<checkboxes.length; i++) {
				checkboxes[i].checked = false;
			}
		}
	}
}

// tui-grid checkbox 기능 추가 class 2
class CheckboxRenderer2 {
	constructor(props) {
        const { grid, rowKey } = props;

        const label = document.createElement('label');
        label.className = 'checkbox';
        label.setAttribute('for', String(rowKey));

        const hiddenInput = document.createElement('input');
        hiddenInput.className = 'hidden-input';
        hiddenInput.id = String(rowKey);

        const customInput = document.createElement('span');
        customInput.className = 'custom-input';

        label.appendChild(hiddenInput);
        label.appendChild(customInput);

        hiddenInput.type = 'checkbox';
        hiddenInput.addEventListener('change', () => {
            if (hiddenInput.checked) {
                grid.check(rowKey);
            } else {
                grid.uncheck(rowKey);
            }
        });

        this.el = label;

        this.render(props);
    }

    getElement() {
        return this.el;
    }

    render(props) {
        const hiddenInput = this.el.querySelector('.hidden-input');
        const checked = Boolean(props.value);

        hiddenInput.checked = checked;
	}
}

// tui-grid rownum 기능 추가 class
class RowNumberRenderer {
	constructor(props) {
        const el = document.createElement('span');
        el.innerHTML = `${props.formattedValue}`;

        this.el = el;
    }

    getElement() {
        return this.el;
    }

    render(props) {
        this.el.innerHTML = `${props.formattedValue}`;
    }
}

// tui-grid button 기능 추가 class
class ButtonRenderer {
	constructor(props) {
        const el = document.createElement('button');
        el.innerHTML = `${props.formattedValue}`;
        el.setAttribute('style', 'cursor:pointer;')
        this.el = el;
    }

    getElement() {
        return this.el;
    }

    render(props) {
        this.el.innerHTML = `${props.formattedValue}`;
    }
}