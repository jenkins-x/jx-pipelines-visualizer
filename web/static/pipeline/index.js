(function(){
    const ansi_up = new AnsiUp;
    const logs = document.getElementById("logs");
    const errors = document.getElementById("errors");
    const downloadLink = document.getElementById("downloadLogs");

    const displayStepsCheckbox = document.querySelector('.display-steps');
    const followLogsCheckbox = document.querySelector('.follow-logs');
    const logsTable = document.querySelector('.logs-table');
    const stickyHeader = document.querySelector('.header-hidden');
    const stickyOption = document.querySelector('.follow-option');

    const cssLineSelected = 'selected-line';

    const addColorThemeOption = () => {
        const themeSwitch = document.querySelector("#theme-switch");
        themeSwitch.addEventListener('click', (e) => {
            if(e.target.checked) {
                logsTable.classList.add('logs-dark-theme');
                localStorage.setItem('logs-dark-theme', true);
            } else {
                logsTable.classList.remove('logs-dark-theme');
                localStorage.removeItem('logs-dark-theme');
            }
        });

        // Init 
        if (localStorage.getItem('logs-dark-theme')) {
            themeSwitch.click();
        }
    };

    const transformLogIntoHtml = (lineNumber, text, type='') => {
        var containerId = "";
        if (text.startsWith('Showing logs for build ')) {
            const regex = /\[32m([^\[])+\[0m/g;
            const matches = text.match(regex);
            if (matches.length == 3) {
                const stage = matches[1].replace('[32m', '').replace('[0m', '').slice(0, -1);
                const container = matches[2].replace('[32m', '').replace('[0m', '').slice(0, -1);
                containerId = "log-" + stage + "-" + container;
            }
        }

        const html = ansi_up.ansi_to_html(text)

        // Transform url to link element
        const transformedText = html.replace(/(https?:\/\/\S+)/g, '<a href="$1">$1</a>');

        return `
        <tr id="logsL${lineNumber}">
            <td class="log-number" data-line-number=${lineNumber}></td>
            <td class="log-line" id="${containerId}">
                <span class="line-text ${type}">${transformedText}</span>
            </td>
        </tr>
        `;
    }

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

    const addStageStepsLinksEvent = () => {
        var links = document.querySelectorAll('.stage-steps-link');

        Array.from(links).forEach(link => {
            link.addEventListener('click', function(event) {
                document.querySelector(link.getAttribute('href')).classList.remove('steps-hidden');
            });
        });
    }

    const addDisplayStepsEvent = () => {
        displayStepsCheckbox.addEventListener('click', function(event) {
            var stages = document.querySelectorAll('.stage');
            Array.from(stages).forEach(stage => {
                if (displayStepsCheckbox.checked) {
                    stage.classList.remove('steps-hidden');
                } else {
                    stage.classList.add('steps-hidden');
                }
            });
        });
    }

    const generateDownloadLink = (logs) => {
        var blob = new Blob([logs], { type : "text/plain;charset=utf-8"});
        downloadUrl = URL.createObjectURL(blob);

        downloadLink.setAttribute("href", downloadUrl);
    }

    const loadByBuildLogUrl = () => {
        fetch(`${LOGS_URL}/logs`).then(response => response.text()).then((response) => {
            logs.innerHTML = transformLogsIntoHtml(response);
            addLinks();
            goToAnchor();
            generateDownloadLink(response);
        }).catch((error)=> {
            errors.innerHTML = transformLogsIntoHtml(error, 'line-error');
        });
    };

    const loadByEventSource = () => {
        const eventSource = new EventSource(`${LOGS_URL}/logs/live`);
        let lineNumber = 0;
        let logsBuffer = "";
        let getAnchor = false;
        let isFinished = false;
        
        downloadLink.remove();

        const repeatOften = () => {
            if(logsBuffer) {
                if(lineNumber === 0) {
                    logs.innerHTML = "";
                }

                logs.insertAdjacentHTML('beforeend', transformLogsIntoHtml(logsBuffer, '', () => ++lineNumber));
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
            errors.innerHTML = transformLogsIntoHtml(e.data, 'line-error');
        });
        eventSource.addEventListener("EOF", function(e) {
            eventSource.close();
            isFinished = true;
        });

        // Waiting the next animation frame to add DOM element
        requestAnimationFrame(repeatOften);   
    };


    // Run
    addStageStepsLinksEvent();
    addDisplayStepsEvent();
    addScrollEvent();
    addColorThemeOption();

    if (BUILD_LOG_URL) {
        loadByBuildLogUrl();
    } else {
        loadByEventSource();
    }
})();