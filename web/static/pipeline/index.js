(function(){
    const ansi_up = new AnsiUp;
    const logs = document.getElementById("logs");
    const errors = document.getElementById("errors");

    const followLogsCheckbox = document.querySelector('.follow-logs');
    const stickyHeader = document.querySelector('.header-hidden');
    const stickyOption = document.querySelector('.follow-options');

    const cssLineSelected = 'selected-line';

    const transformLogIntoHtml = (lineNumber, text, type='') =>
        `<tr id="logsL${lineNumber}">
            <td class="log-number" data-line-number=${lineNumber}></td>
            <td class="log-line">
                <span class="line-text ${type}">${text}</span>
            </td>
        </tr>`;

    const transformLogsIntoHtml = (logsString, type='', givenIndex) =>
        logsString
            .split('\n')
            .slice(1, -1)
            .map((line, index) => transformLogIntoHtml(givenIndex ? givenIndex() : index+1, line, type))
            .join('\n');

    const onClickLineNumber = event => {
        const elem = event.target;

        if (location.hash) {
            const previousClicked = document.querySelector(location.hash);
            previousClicked.classList.remove(cssLineSelected);
        }

        history.pushState(null, null, `#logsL${elem.dataset.lineNumber}`);
        elem.parentElement.classList.add(cssLineSelected);
    };

    const addLinks = () => 
        document.querySelectorAll('.log-number').forEach(elem => elem.addEventListener('click', onClickLineNumber));

    const goToAnchor = () => {
        if (location.hash) {
            const elem = document.querySelector(location.hash);
            if (elem) {
                elem.scrollIntoView({block: 'center', inline: 'center', behavior: 'smooth'});
                elem.classList.add(cssLineSelected);
                return true;
            }
        }
        return false;
    };

    const addScrollEvent = () => {
        window.addEventListener('scroll', function(e) {
            if (window.scrollY > 300) {
                stickyHeader.classList.add('sticky-header');
                stickyOption.classList.add('sticky-option')
            } else if (window.scrollY < 300) {
                stickyHeader.classList.remove('sticky-header');
                stickyOption.classList.remove('sticky-option')
            }
        });
    }

    const loadByBuildLogUrl = () => {
        fetch(`${LOGS_URL}/logs`).then(response => response.text()).then((response) => {
            console.log(response);
            logs.innerHTML = transformLogsIntoHtml(ansi_up.ansi_to_html(response));
            addLinks();
            goToAnchor();
        }).catch((error)=> {
            errors.innerHTML = transformLogsIntoHtml(ansi_up.ansi_to_html(error), 'line-error');
        });
    };

    const loadByEventSource = () => {
        const eventSource = new EventSource(`${LOGS_URL}/logs/live`);
        let lineNumber = 0;
        let logsBuffer = "";
        let getAnchor = false;
        let isFinished = false;

        const repeatOften = () => {
            if(logsBuffer) {
                logs.insertAdjacentHTML('beforeend', transformLogsIntoHtml(ansi_up.ansi_to_html(logsBuffer), '', () => ++lineNumber));
                addLinks();
                if (!getAnchor) {
                    getAnchor = goToAnchor();
                }
                if(followLogsCheckbox.checked) {
                    const lastLog = document.getElementById(`logsL${lineNumber}`);
                    lastLog.scrollIntoView({block: 'end', inline: 'end', behavior: 'smooth'});
                }
                logsBuffer = "";
            }
            if(!isFinished) {
                requestAnimationFrame(repeatOften);
            }
        };
        
        eventSource.addEventListener("log", function(e) {
            logsBuffer += e.data + "\n";
        }, {passive: true});
        eventSource.addEventListener("error", function(e) {
            errors.innerHTML = transformLogsIntoHtml(ansi_up.ansi_to_html(e.data), 'line-error');
        });
        eventSource.addEventListener("EOF", function(e) {
            eventSource.close();
            isFinished = true;
        });

        // Waiting the next animation frame to add DOM element
        requestAnimationFrame(repeatOften);   
    };


    // Run

    addScrollEvent();

    if (BUILD_LOG_URL) {
        loadByBuildLogUrl();
    } else {
        loadByEventSource();
    }
})();