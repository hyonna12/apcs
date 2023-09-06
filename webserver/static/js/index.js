// 팝업 스크롤 막기
function scrollLock() {
    const body = document.getElementsByTagName('body')[0];
    body.classList.add('scrolllock')
    
}

// 팝업 스크롤 막기 해제
function scrollLockOff() {
    const body = document.getElementsByTagName('body')[0];
    body.classList.remove('scrolllock')
}

$(document).ready(function(){
    // tab menu
    $('.tab-menu ul li').on('click', function() {
        const num = $('.tab-menu ul li').index($(this));
        
        $('.tab-menu ul li').removeClass('active');
        $('.tab-con .tab-con-div').removeClass('on');

        $('.tab-menu li:eq(' + num + ')').addClass('active');
        $('.tab-con > .tab-con-div:eq(' + num + ')').addClass('on');
        
    });

     // list chk all
    $('#list_chkAll').click(function() {
        if($('#list_chkAll').is(':checked')) $('input[name=list_chk]').prop('checked', true);
        else $('input[name=list_chk]').prop('checked', false);

        $('input[name=list_chk]').click(function() {
            var total = $('input[name=list_chk]').length;
            var checked = $('input[name=list_chk]:checked').length;

            if(total != checked) $('#list_chkAll').prop('checked', false);
            else $('#list_chkAll').prop('checked', true);
        })
    })

    // 숫자만 입력
	$('input[name=number_input]').keyup(function(){
        $(this).val(Number($(this).val().replace(/[^0-9]/g,"")));
    })

})