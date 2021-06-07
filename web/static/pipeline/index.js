(function(){
    const ansi_up = new AnsiUp;
    const logs = document.getElementById("logs");
    const downloadLink = document.getElementById("downloadLogs");

    const followLogsCheckbox = document.querySelector('.follow-logs');
    const logsTable = document.querySelector('.logs-table');
    const stickyHeader = document.querySelector('.header-hidden');
    const stickyOption = document.querySelector('.follow-option');

    const cssLineSelected = 'selected-line';

    let currentStep = '';
    let allIsOpen = false;

    // Utils

    const getAllParentSteps = () => document.querySelectorAll('tr[data-is-parent-step=true]');
    const getParentStep = name => document.querySelector(`tr[data-step=${name}][data-is-parent-step=true]`);

    const goToAnchor = () => {
        if (location.hash) {
            const elem = document.querySelector(location.hash);
            if (elem) {
                if(location.hash.includes('logsL')) {
                    // We open the parent before
                    const stepParent = getParentStep(elem.dataset.step);
                    toggleStep(stepParent, true);
                    elem.scrollIntoView({block: 'center', inline: 'center', behavior: 'smooth'});
                    elem.classList.add(cssLineSelected);
                    return true;
                }
            }
        }
        return false;
    };

    const toggleClassName = (selector, cssClass) => {
        const element = document.querySelector(selector);
        if (element.classList.contains(cssClass)) {
            element.classList.remove(cssClass);
        } else {
            element.classList.add(cssClass);
        }
    };

    const toggleStep = (stepElement, toOpen = null) => {
        const stepName = stepElement.dataset.step;

        if ((stepElement.dataset.open === 'true' && toOpen === null)  || toOpen === false) {
            stepElement.dataset.open = false;
            document.querySelectorAll(`tr[data-step=${stepName}]`).forEach(childStep => childStep.classList.add('step-line-hidden'));
        } else {
            stepElement.dataset.open = true;
            document.querySelectorAll(`tr[data-step=${stepName}]`).forEach(childStep => childStep.classList.remove('step-line-hidden'));
        }
    };

    // Listeners

    const addClickEventToStep = () => document.getElementById('toggle-steps').addEventListener('click', toggleAllSteps);
    const addClickOpenTrace = () => {
        const elem = document.getElementById('open-trace');
        if (elem) {
            elem.addEventListener('click', openTrace);
        }
    };
    const addClickShowTimeline = () => document.getElementById('show-timeline').addEventListener('click', showTimeline);
    const addLinks = () => document.querySelectorAll('.log-number').forEach(elem => elem.addEventListener('click', onClickLineNumber));

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
    };

    const addStageStepsLinksEvent = () => {
        const openShowTimeline = () => {
            document.querySelector('#pipeline-timeline').classList.remove('steps-hidden');
        };
        document.querySelectorAll('.stages .stage-steps-link').forEach(link => link.addEventListener('click', openShowTimeline));

        document.querySelectorAll('.link-to-console').forEach(link => {
            link.addEventListener('click', () => {
                const stepToOpen = document.querySelector(link.getAttribute('href'));
                toggleStep(stepToOpen.parentElement, true);
            });
        });
    };

    // Options

    const openTrace = () => {
        const elem = document.getElementById('open-trace');
        if (elem) {
            const traceURL = elem.getAttribute('href');
            window.open(traceURL);
        }
    };

    const showTimeline = () => toggleClassName('#pipeline-timeline', 'steps-hidden');

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

    const onClickParentStep = event => toggleStep(event.currentTarget);
    
    const onClickLineNumber = event => {
        const elem = event.target;

        if (location.hash) {
            const previousClicked = document.querySelector(location.hash);
            previousClicked.classList.remove(cssLineSelected);
        }

        history.pushState(null, null, `#logsL${elem.dataset.lineNumber}`);
        elem.parentElement.classList.add(cssLineSelected);
    };

    const toggleAllSteps = () => {
        allIsOpen = !allIsOpen;
        getAllParentSteps().forEach(step => toggleStep(step, allIsOpen));
    };

    const generateDownloadLink = (logs) => {
        var blob = new Blob([logs], { type : "text/plain;charset=utf-8"});
        downloadUrl = URL.createObjectURL(blob);

        downloadLink.setAttribute("href", downloadUrl);
    };

    // Fetch + Read + Enhance logs

    const updateStepStatusIcon = () => {
        const currentRunning = document.querySelector('.log-status-icon[data-status=Running]');
        if(currentRunning) {
            const containerId = "log-" + currentRunning.parentElement.dataset.step;
            currentRunning.dataset.status = STEPS[containerId] ? STEPS[containerId].status : '';
        }
        if(currentStep){
            document.querySelector(`tr[data-step=${currentStep}][data-is-parent-step=true] .log-status-icon`).dataset.status = 'Running';
        }
    };

    const transformLogIntoHtml = (lineNumber, text, type='') => {
        let containerId = '';
        if (text.startsWith('Showing logs for build ')) {
            const regex = /\[32m([^\[])+\[0m/g;
            const matches = text.match(regex);
            if (matches.length == 3) {
                const stage = matches[1].replace('[32m', '').replace('[0m', '').slice(0, -1);
                const container = matches[2].replace('[32m', '').replace('[0m', '').slice(0, -1);
                containerId = "log-" + stage + "-" + container;
                currentStep = stage + "-" + container;
            }
        }

        let cssClass = 'step-line-hidden';
        if (type === 'line-error') {
            cssClass = 'step-line';
        }

        const html = ansi_up.ansi_to_html(text)

        // Transform url to link element
        const transformedText = html.replace(/(https?:\/\/\S+)/g, '<a href="$1">$1</a>');

        return `
        <tr id="logsL${lineNumber}" data-step="${currentStep}" data-is-parent-step="${containerId !== ''}" class="${cssClass}">
            <td class="log-number" data-line-number="${lineNumber}"></td>
            <td class="log-dropdown-icon"></td>
            <td class="log-status-icon" data-status="${STEPS[containerId] ? STEPS[containerId].status : ''}"></td>
            <td class="log-timer">${STEPS[containerId] ? STEPS[containerId].timer : ''}</td>
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

    const loadByBuildLogUrl = () => {
        const hostnameWithoutLogin = window.location.origin.replace(/\/\/[^@]*@/, '//');

        fetch(`${hostnameWithoutLogin}${LOGS_URL}/logs`).then((response) => {
            if (response.status == 404) {
                throw new Error('Archived logs not found in the long term storage');
            }
            if (!response.ok) {
                throw new Error('Failed to retrieve the archived logs from the long term storage: ' + response.status + ' ' + response.statusText);
            }
            return response.text();
        }).then((response) => {
            logs.innerHTML = transformLogsIntoHtml(response);
            addLinks();
            goToAnchor();
            generateDownloadLink(response);
            getAllParentSteps().forEach(parentStep => parentStep.addEventListener('click', onClickParentStep));
        }).catch((error) => {
            logs.innerHTML = transformLogIntoHtml(0, error.toString(), 'line-error');
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
                getAllParentSteps().forEach(parentStep => parentStep.addEventListener('click', onClickParentStep));
                
                if (!getAnchor) {
                    getAnchor = goToAnchor();
                }
                
                if(followLogsCheckbox.checked) {
                    const lastLog = document.getElementById(`logsL${lineNumber}`);
                    getAllParentSteps().forEach(step => toggleStep(step, false));
                    if(currentStep) {
                        // Open current step
                        toggleStep(getParentStep(currentStep), true);
                    }
                    // Update step status icon
                    updateStepStatusIcon();
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
            logs.innerHTML = transformLogIntoHtml(0, e.data, 'line-error');
        });
        eventSource.addEventListener("EOF", function(e) {
            eventSource.close();
            isFinished = true;
        });

        // Waiting the next animation frame to add DOM element
        requestAnimationFrame(repeatOften);   
    };

    // Init

    const init = () => {
        addScrollEvent();
        addColorThemeOption();
        addClickEventToStep();
     
        if (!ARCHIVE) {
            addStageStepsLinksEvent();
            addClickShowTimeline();
            addClickOpenTrace();
        }
    
        if (BUILD_LOG_URL) {
            loadByBuildLogUrl();
        } else {
            loadByEventSource();
        }
    }

    // Run
   init();
})();