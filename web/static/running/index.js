(function(){
    const dataTable = $('#dataTable').DataTable({
        lengthMenu: [ [10, 25, 50, 100, -1], [10, 25, 50, 100, "All"] ],
        pageLength: 25,
        order: [
            [0, 'desc'],
            [1, 'desc'],
            [2, 'desc']
        ],
        columnDefs: [
            { targets: 'start', visible: false }
        ]
    });

    const loadByEventSource = () => {
        const eventSource = new EventSource(`/running/events`);
        
        eventSource.addEventListener("added", function(e) {
            let pipeline = JSON.parse(e.data);
            var row = dataTable.row.add([
                '<a href="/' + pipeline.Owner + '/' + pipeline.Repository + '">' + pipeline.Owner + '/' + pipeline.Repository + '</a>',
                '<a href="/' + pipeline.Owner + '/' + pipeline.Repository + '/' + pipeline.Branch + '">' + pipeline.Branch + '</a>',
                '<a href="/' + pipeline.Owner + '/' + pipeline.Repository + '/' + pipeline.Branch + '/' + pipeline.Build + '">' + pipeline.Build + '</a>',
                pipeline.Context,
                pipeline.Stage,
                pipeline.Step,
                moment.duration(moment().diff(moment(pipeline.StepStartTime))).seconds() + 's',
                pipeline.StepStartTime
            ]);
            row.node().setAttribute('id', pipeline.Name + '-' + pipeline.Stage.replaceAll(' ', '-') + '-' + pipeline.Step.replaceAll(' ', '-'));
            row.draw();
        }, {passive: true});
        eventSource.addEventListener("deleted", function(e) {
            let pipeline = JSON.parse(e.data);
            let id = '#' + pipeline.Name + '-' + pipeline.Stage.replaceAll(' ', '-') + '-' + pipeline.Step.replaceAll(' ', '-');
            dataTable.row(id).remove().draw();
        }, {passive: true});
    };

    const refreshDuration = () => {
        dataTable.rows().every( function ( rowIdx, tableLoop, rowLoop ) {
            let startTime = dataTable.cell({row: rowIdx, column: 7}).data();
            var duration = moment.duration(moment().diff(moment(startTime)));
            if (duration.asSeconds() < 60) {
                dataTable.cell({row: rowIdx, column: 6}).data(duration.seconds() + 's');
            } else {
                dataTable.cell({row: rowIdx, column: 6}).data(duration.minutes() + 'm' + duration.seconds() + 's');
            }
        });
    }

    // Init

    const init = () => {
        loadByEventSource();
        refreshDuration();
        setInterval(refreshDuration, 1000);
    }

    // Run
    init();
})();