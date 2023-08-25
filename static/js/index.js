$(document).ready(function() {
    // header active
    $('.main-header .header-area .menu-box').mouseenter(function() {
        $('.main-header .header-wrap').addClass('active');
        $('.main-header .hidden-layer').stop().slideDown('fast', function() {
            $('.main-header .lnb').stop().fadeIn('fast');
        });

        if(!$('.main-header .header-wrap').hasClass('active')){
            $('.main-header .header-wrap').addClass('active')
        }
    });

    $('.main-header .header-area .menu-box').mouseleave(function() {
        $('.main-header .header-wrap').removeClass('active');
        $('.main-header .lnb').stop().fadeOut('fast', function() {
            $('.main-header .hidden-layer').stop().slideUp('fast');
        });
        
        if($('.main-hedaer .header-wrap').hasClass('active')) {
            $('.main-header .header-wrap').removeClass('active');
        }
    })

    // sub header active
    $('.sub-header .header-area .menu-box').mouseenter(function() {
        $('.sub-header .header-wrap').addClass('active');
        $('.sub-header .hidden-layer').stop().slideDown('fast', function() {
            $('.sub-header .lnb').stop().fadeIn('fast');
        });

        if(!$('.sub-header .header-wrap').hasClass('active')){
            $('.sub-header .header-wrap').addClass('active')
        }
    });

    $('.sub-header .header-area .menu-box').mouseleave(function() {
        $('.sub-header .header-wrap').removeClass('active');
        $('.sub-header .lnb').stop().fadeOut('fast', function() {
            $('.sub-header .hidden-layer').stop().slideUp('fast');
        });

        if($('.sub-hedaer .header-wrap').hasClass('active')) {
            $('.sub-header .header-wrap').removeClass('active');
        }
    })

    // sub2.html tab menu
    $('.tab-menu ul li').on('click', function() {
        const num = $('.tab-menu ul li').index($(this));
        
        $('.tab-menu ul li').removeClass('active');
        $('.tab-con > div').removeClass('on');

        $('.tab-menu li:eq(' + num + ')').addClass('active');
        $('.tab-con > div:eq(' + num + ')').addClass('on');
    });

    // company list chk all
    $('#company_list_chkAll').click(function() {
        if($('#company_list_chkAll').is(':checked')) $('input[name=company_chk]').prop('checked', true);
        else $('input[name=company_chk]').prop('checked', false);

        $('input[name=company_chk]').click(function() {
            var total = $('input[name=company_chk]').length;
            var checked = $('input[name=company_chk]:checked').length;

            if(total != checked) $('#company_list_chkAll').prop('checked', false);
            else $('#company_list_chkAll').prop('checked', true);
        })
    })

    $('#company_list_chkAll').click(function() {
         
    })

    // group list chk all
    $('#group_list_chkAll').click(function() {
        if($('#group_list_chkAll').is(':checked')) $('input[name=group_chk]').prop('checked', true);
        else $('input[name=group_chk]').prop('checked', false);

        $('input[name=group_chk]').click(function() {
            var total = $('input[name=group_chk]').length;
            var checked = $('input[name=group_chk]:checked').length;

            if(total != checked) $('#group_list_chkAll').prop('checked', false);
            else $('#group_list_chkAll').prop('checked', true);
        })
    })

    // group > company list chk all
    $('#group_company_chkAll').click(function() {
        if($('#group_company_chkAll').is(':checked')) $('input[name=group_company_chk]').prop('checked', true);
        else $('input[name=group_company_chk]').prop('checked', false);

        $('input[name=group_company_chk]').click(function() {
            var total = $('input[name=group_company_chk]').length;
            var checked = $('input[name=group_company_chk]:checked').length;

            console.log(total,'to')
            console.log(checked)

            if(total != checked) $('#group_company_chkAll').prop('checked', false);
            else $('#group_company_chkAll').prop('checked', true);
        })
    })

    // email direct
    $(function() {
        $('#selboxDirect').hide();
        
        $('#selbox_email').change(function() {
            if($('#selbox_email').val() == 'direct') {
                $('#selboxDirect').show();
            } else $('#selboxDirect').hide();
        })
    })

    // file name
    $(function() {
        var fileTarget = $('.file-input');

        fileTarget.on('change', function() {
            if(window.FileReader) {
                var filename = $(this)[0].files[0].name;
            }
            else { // IE version
                var filename = $(this).val().split('/').pop().split('\\').pop();
            }

            $(this).siblings('.file-name').val(filename);
        });
    });

    // mobile side menu
    $('.menu').on('click', function() {
        $('.side-wrap').addClass('active')
    })

    $('.menu-close').on('click', function(){
        $('.side-wrap').removeClass('active')
    })

    $('.side-layer').on('click', function() {
        $('.side-wrap').removeClass('active')
    })

    // 사이드 2차메뉴
    $('.side-menu-wrap .menu-box > li').on('click', function(){
        const num = $('.side-menu-wrap .menu-box > li').index($(this));

        $('.side-menu-wrap .menu-box > li').removeClass('active');

        $('.side-menu-wrap .menu-box > li:eq(' + num + ')').toggleClass('active');
        $('.side-menu-wrap ')
    })
})

