$(document).ready(() => {
    $.fn.dataTable.ext.order['dom-order'] = function( _, col ) { 
        return this.api().column( col, {order:'index'} ).nodes().map(td => $(td).data('order'));
    };

    $('#dataTable').DataTable({
        lengthMenu: [ [10, 25, 50, 100, -1], [10, 25, 50, 100, "All"] ],
        pageLength: 25,
        order: [[5, 'desc']],
        columnDefs: [
            { targets: 'branch', orderDataType: 'dom-order' },
            { targets: 'start', orderDataType: 'dom-order' },
            { targets: 'end', orderDataType: 'dom-order' },
            { targets: 'duration', orderDataType: 'dom-order', type: 'numeric' },
            { targets: 'author', visible: false }
        ]
    });
});